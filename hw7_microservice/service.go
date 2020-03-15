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
	ACL      map[string][]string
	Logs     chan *Event
	Stats    chan *Event
	LogSubs  map[chan *Event]struct{}
	StatSubs map[chan *Event]struct{}
	mutex    *sync.RWMutex
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
	ch := srv.newLogSub()
	for c := range ch {
		err := server.Send(c)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("log send error: ", err.Error())
			break
		}
	}
	return nil
}

func (srv *AdminManager) newLogSub() chan *Event {
	ch := make(chan *Event, 2)
	srv.mutex.Lock()
	srv.LogSubs[ch] = struct{}{}
	srv.mutex.Unlock()
	return ch
}

func (srv *AdminManager) Statistics(in *StatInterval, server Admin_StatisticsServer) error {
	ticker := time.NewTicker(time.Second * time.Duration(in.IntervalSeconds))
	defer ticker.Stop()
	st := &Stat{
		Timestamp:  time.Now().Unix(),
		ByMethod:   make(map[string]uint64),
		ByConsumer: make(map[string]uint64),
	}
	ch := srv.newLogSub()

	for {
		select {
		case <-ticker.C:
			err := server.Send(st)
			if err == io.EOF {
				return nil
			}
			if err != nil {
				fmt.Println("stat send error: ", err.Error())
				return nil
			}
			st.ByMethod = make(map[string]uint64)
			st.ByConsumer = make(map[string]uint64)
			st.Timestamp = time.Now().Unix()
		case v := <-ch:
			if _, ok := st.ByMethod[v.Method]; ok {
				st.ByMethod[v.Method]++
			} else {
				st.ByMethod[v.Method] = 1
			}
			if _, ok := st.ByConsumer[v.Consumer]; ok {
				st.ByConsumer[v.Consumer]++
			} else {
				st.ByConsumer[v.Consumer] = 1
			}

		case <-server.Context().Done():
			return nil
		}
	}

	return nil
}

func (srv *AdminManager) newStatSub() chan *Event {
	ch := make(chan *Event, 2)
	srv.mutex.Lock()
	srv.StatSubs[ch] = struct{}{}
	srv.mutex.Unlock()
	return ch
}

func StartMyMicroservice(ctx context.Context, listenAddr, ACLData string) error {
	var acl map[string][]string
	err := json.Unmarshal([]byte(ACLData), &acl)
	if err != nil {
		return err
	}

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(aclInterceptor),
		grpc.StreamInterceptor(streamACLInterceptor),
	)

	store := &Storage{
		ACL:     acl,
		Logs:    make(chan *Event, 2),
		Stats:   make(chan *Event, 2),
		mutex:   &sync.RWMutex{},
		LogSubs: make(map[chan *Event]struct{}),
	}

	RegisterBizServer(server, &BizManager{store})
	RegisterAdminServer(server, &AdminManager{store})
	fmt.Println("starting server at ", listenAddr)

	// TODO: needs err handling https://www.atatus.com/blog/goroutines-error-handling/
	go server.Serve(lis)

	go func() {
		for val := range store.Logs {
			store.mutex.RLock()
			for c := range store.LogSubs {
				c <- val
			}
			store.mutex.RUnlock()
			store.mutex.RLock()
			for c := range store.StatSubs {
				c <- val
			}
			store.mutex.RUnlock()
		}
		store.mutex.Lock()
		for ch := range store.LogSubs {
			close(ch)
			delete(store.LogSubs, ch)
		}
		store.LogSubs = nil
		store.mutex.Unlock()
	}()

	go func() {
		<-ctx.Done()
		close(store.Logs)
		close(store.Stats)
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
	return handler(ctx, req)
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
	return handler(srv, ss)
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
