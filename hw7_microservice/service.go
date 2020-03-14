package main

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
)

type BizManager struct {
	*Storage
}

type AdminManager struct {
	*Storage
}

type Storage struct {
	ACL  map[string][]string
	Logs chan *Event
	//Stats    chan *Stat
	LogSubs map[chan *Event]struct{}
	//StatSubs map[chan *Event]struct{}
	mutex    sync.Mutex
	capacity int
}

func (srv *BizManager) Check(ctx context.Context, in *Nothing) (*Nothing, error) {
	return &Nothing{Dummy: true}, nil
}
func (srv *BizManager) Add(ctx context.Context, in *Nothing) (*Nothing, error) {
	return &Nothing{Dummy: true}, nil
}
func (srv *BizManager) Test(ctx context.Context, in *Nothing) (*Nothing, error) {
	return &Nothing{Dummy: true}, nil
}

func (srv *AdminManager) Logging(in *Nothing, server Admin_LoggingServer) error {
	ch := make(chan *Event, 1)
	srv.mutex.Lock()
	srv.LogSubs[ch] = struct{}{}
	srv.mutex.Unlock()

	go func() {
		for c := range ch {
			err := server.Send(c)
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println(err)
				break
			}
		}
		//close(ch)
		//srv.mutex.Lock()
		//delete(srv.LogSubs, ch)
		//srv.mutex.Unlock()
		//return
	}()
	return nil
}

func (srv *AdminManager) Statistics(in *StatInterval, server Admin_StatisticsServer) error {
	return nil
}

func StartMyMicroservice(ctx context.Context, listenAddr, ACLData string) error {
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}

	var acl map[string][]string
	err = json.Unmarshal([]byte(ACLData), &acl)
	if err != nil {
		return err
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(aclInterceptor),
		grpc.StreamInterceptor(streamACLInterceptor),
	)

	store := &Storage{
		ACL:      acl,
		Logs:     make(chan *Event, 20),
		capacity: 1,
		mutex:    sync.Mutex{},
		LogSubs:  make(map[chan *Event]struct{}),
	}
	go func() {
		for val := range store.Logs {
			store.mutex.Lock()
			for c := range store.LogSubs {
				c <- val
			}
			store.mutex.Unlock()
		}
		store.mutex.Lock()
		for c := range store.LogSubs {
			close(c)
			delete(store.LogSubs, c)
		}
		store.LogSubs = nil
		store.mutex.Unlock()

	}()
	RegisterBizServer(server, &BizManager{store})
	RegisterAdminServer(server, &AdminManager{store})
	fmt.Println("starting server at ", listenAddr)

	// TODO: needs err handling https://www.atatus.com/blog/goroutines-error-handling/
	go func() {
		server.Serve(lis)
	}()

	go func() {
		<-ctx.Done()
		close(store.Logs)
		//close(store.Stats)
		server.Stop()
		lis.Close()
	}()

	return nil
}

func aclInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	c := md.Get("consumer")
	p, _ := peer.FromContext(ctx)
	if c == nil {
		return nil, status.Errorf(codes.Unauthenticated, "empty consumer")
	}

	consumer := c[0]
	method := info.FullMethod
	manager := info.Server.(*BizManager)

	valid := validate(manager.ACL, consumer, method)

	if !valid {
		return nil, status.Errorf(codes.Unauthenticated, "method %s for consumer %s is not allowed", method, consumer)
	}

	manager.Logs <- &Event{Timestamp: time.Now().Unix(), Consumer: consumer, Method: method, Host: p.Addr.String()}

	reply, err := handler(ctx, req)

	return reply, err
}

func streamACLInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	md, _ := metadata.FromIncomingContext(ss.Context())
	c := md.Get("consumer")
	p, _ := peer.FromContext(ss.Context())
	if c == nil {
		return status.Errorf(codes.Unauthenticated, "empty consumer")
	}

	consumer := c[0]
	method := info.FullMethod
	manager := srv.(*AdminManager)

	valid := validate(manager.ACL, consumer, method)

	if !valid {
		return status.Errorf(codes.Unauthenticated, "method %s for consumer %s is not allowed", method, consumer)
	}

	manager.Logs <- &Event{Timestamp: time.Now().Unix(), Consumer: consumer, Method: method, Host: p.Addr.String()}
	// handle err
	err := handler(srv, ss)

	return err
}

func validate(acl map[string][]string, consumer string, method string) bool {
	valid := false
	methodSplit := strings.Split(method, "/")
	for _, v := range acl[consumer] {
		if v == method {
			valid = true
			continue
		}
		if strings.Contains(v, methodSplit[1]) {
			if strings.HasSuffix(v, "*") || strings.HasSuffix(v, methodSplit[2]) {
				valid = true
			}
		}
	}
	return valid
}
