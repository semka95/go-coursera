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
	"strconv"
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
	ACL       map[string][]string
	Events    chan *Event
	EventSubs []chan *Event
	mutex     *sync.RWMutex
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
	md, _ := metadata.FromIncomingContext(server.Context())
	id, _ := strconv.Atoi(md.Get("subID")[0])
	srv.mutex.Lock()
	ch := srv.EventSubs[id]
	srv.mutex.Unlock()
	for c := range ch {
		err := server.Send(c)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (srv *AdminManager) Statistics(in *StatInterval, server Admin_StatisticsServer) error {
	ticker := time.NewTicker(time.Second * time.Duration(in.IntervalSeconds))
	defer ticker.Stop()
	st := Stat{
		Timestamp:  time.Now().Unix(),
		ByMethod:   make(map[string]uint64),
		ByConsumer: make(map[string]uint64),
	}
	md, _ := metadata.FromIncomingContext(server.Context())
	id, _ := strconv.Atoi(md.Get("subID")[0])
	srv.mutex.Lock()
	ch := srv.EventSubs[id]
	srv.mutex.Unlock()

	for {
		select {
		case <-ticker.C:
			err := server.Send(&st)
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return nil
			}
			st.ByMethod = make(map[string]uint64)
			st.ByConsumer = make(map[string]uint64)
			st.Timestamp = time.Now().Unix()
		case v, ok := <-ch:
			if !ok {
				ch = nil
				return nil
			}
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
		grpc.UnaryInterceptor(unaryInterceptor),
		grpc.StreamInterceptor(streamInterceptor),
	)

	store := &Storage{
		ACL:       acl,
		Events:    make(chan *Event, 100),
		mutex:     &sync.RWMutex{},
		EventSubs: []chan *Event{},
	}

	RegisterBizServer(server, &BizManager{store})
	RegisterAdminServer(server, &AdminManager{store})
	fmt.Println("starting server at ", listenAddr)

	go server.Serve(lis)

	go func() {
		for val := range store.Events {
			store.mutex.RLock()
			for _, c := range store.EventSubs {
				c <- val
			}
			store.mutex.RUnlock()
		}
		store.mutex.Lock()
		for _, ch := range store.EventSubs {
			close(ch)
		}
		store.EventSubs = nil
		store.mutex.Unlock()
	}()

	go func() {
		<-ctx.Done()
		close(store.Events)
		server.Stop()
		lis.Close()
	}()

	return nil
}

func unaryInterceptor(
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

	manager.Events <- &Event{Timestamp: time.Now().Unix(), Consumer: consumer, Method: method, Host: p.Addr.String()}
	reply, err := handler(ctx, req)
	return reply, err
}

func (srv *AdminManager) newSub() (chan *Event, int) {
	ch := make(chan *Event, 2)
	srv.mutex.Lock()
	key := len(srv.EventSubs)
	srv.mutex.Unlock()
	srv.mutex.Lock()
	srv.EventSubs = append(srv.EventSubs, ch)
	srv.mutex.Unlock()
	<-ch
	return ch, key
}

func streamInterceptor(
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

	manager.Events <- &Event{Timestamp: time.Now().Unix(), Consumer: consumer, Method: method, Host: p.Addr.String()}
	_, id := manager.newSub()
	md.Set("subID", strconv.Itoa(id))
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
