package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

type Server struct {
	ctx       context.Context
	ACL       map[string][]string
	Events    chan *Event
	EventSubs map[int]chan *Event
	mutex     *sync.RWMutex
}

func (s *Server) Check(ctx context.Context, in *Nothing) (*Nothing, error) {
	return &Nothing{Dummy: true}, nil
}
func (s *Server) Add(ctx context.Context, in *Nothing) (*Nothing, error) {
	return &Nothing{Dummy: true}, nil
}
func (s *Server) Test(ctx context.Context, in *Nothing) (*Nothing, error) {
	return &Nothing{Dummy: true}, nil
}

func (s *Server) Logging(in *Nothing, server Admin_LoggingServer) error {
	md, _ := metadata.FromIncomingContext(server.Context())
	id, _ := strconv.Atoi(md.Get("subID")[0])
	s.mutex.RLock()
	ch := s.EventSubs[id]
	s.mutex.RUnlock()

	for {
		select {
		case c := <-ch:
			err := server.Send(c)
			if err != nil {
				return err
			}
		case <-s.ctx.Done():
			return nil
		}
	}

	return nil
}

func (s *Server) Statistics(in *StatInterval, server Admin_StatisticsServer) error {
	ticker := time.NewTicker(time.Second * time.Duration(in.IntervalSeconds))
	defer ticker.Stop()

	st := Stat{
		Timestamp:  time.Now().Unix(),
		ByMethod:   make(map[string]uint64),
		ByConsumer: make(map[string]uint64),
	}

	md, _ := metadata.FromIncomingContext(server.Context())
	id, _ := strconv.Atoi(md.Get("subID")[0])
	s.mutex.RLock()
	ch := s.EventSubs[id]
	s.mutex.RUnlock()

	for {
		select {
		case <-ticker.C:
			st.Timestamp = time.Now().Unix()
			err := server.Send(&st)
			if err != nil {
				return err
			}
			st.ByMethod = make(map[string]uint64)
			st.ByConsumer = make(map[string]uint64)
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
		case <-s.ctx.Done():
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

	s := &Server{
		ctx:       ctx,
		ACL:       acl,
		Events:    make(chan *Event, 10),
		mutex:     &sync.RWMutex{},
		EventSubs: make(map[int]chan *Event),
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(s.unaryInterceptor),
		grpc.StreamInterceptor(s.streamInterceptor),
	)

	RegisterBizServer(server, s)
	RegisterAdminServer(server, s)
	log.Println("starting server at ", listenAddr)

	go server.Serve(lis)

	go func() {
		for {
			select {
			case val := <-s.Events:
				s.mutex.RLock()
				for _, c := range s.EventSubs {
					c <- val
				}
				s.mutex.RUnlock()
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		<-ctx.Done()
		s.mutex.Lock()
		for _, ch := range s.EventSubs {
			close(ch)
		}
		s.EventSubs = nil
		s.mutex.Unlock()
		close(s.Events)
		server.Stop()
		lis.Close()
	}()

	return nil
}

func (s *Server) unaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	p, _ := peer.FromContext(ctx)

	c := md.Get("consumer")
	if c == nil {
		return nil, status.Errorf(codes.Unauthenticated, "empty consumer")
	}
	consumer := c[0]

	method := info.FullMethod

	valid := validate(s.ACL, consumer, method)
	if !valid {
		return nil, status.Errorf(codes.Unauthenticated, "method %s for consumer %s is not allowed", method, consumer)
	}

	s.Events <- &Event{Timestamp: time.Now().Unix(), Consumer: consumer, Method: method, Host: p.Addr.String()}

	return handler(ctx, req)
}

func (s *Server) streamInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	md, _ := metadata.FromIncomingContext(ss.Context())
	p, _ := peer.FromContext(ss.Context())

	c := md.Get("consumer")
	if c == nil {
		return status.Errorf(codes.Unauthenticated, "empty consumer")
	}
	consumer := c[0]

	method := info.FullMethod

	valid := validate(s.ACL, consumer, method)
	if !valid {
		return status.Errorf(codes.Unauthenticated, "method %s for consumer %s is not allowed", method, consumer)
	}

	s.Events <- &Event{Timestamp: time.Now().Unix(), Consumer: consumer, Method: method, Host: p.Addr.String()}

	ch := make(chan *Event, 2)
	s.mutex.Lock()
	key := len(s.EventSubs)
	s.EventSubs[key] = ch
	s.mutex.Unlock()
	<-ch
	md.Set("subID", strconv.Itoa(key))

	return handler(srv, ss)
}

func validate(acl map[string][]string, consumer string, method string) bool {
	valid := false
	methodSplit := strings.Split(method, "/")
	for _, v := range acl[consumer] {
		if v == method {
			valid = true
			break
		}
		if strings.Contains(v, methodSplit[1]) {
			if strings.HasSuffix(v, "*") || strings.HasSuffix(v, methodSplit[2]) {
				valid = true
			}
		}
	}
	return valid
}
