package main

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net"
	"strings"

	"google.golang.org/grpc"
)

type BizManager struct {
	ACL map[string][]string
}

type AdminManager struct {
	ACL map[string][]string
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
	RegisterBizServer(server, &BizManager{ACL: acl})
	RegisterAdminServer(server, &AdminManager{ACL: acl})
	fmt.Println("starting server at ", listenAddr)

	// TODO: needs err handling https://www.atatus.com/blog/goroutines-error-handling/
	go func() {
		server.Serve(lis)
	}()

	go func() {
		<-ctx.Done()
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
	if c == nil {
		return nil, status.Errorf(codes.Unauthenticated, "empty consumer")
	}

	consumer := c[0]
	method := info.FullMethod
	acl := info.Server.(*BizManager).ACL

	valid := validate(acl, consumer, method)

	if !valid {
		return nil, status.Errorf(codes.Unauthenticated, "method %s for consumer %s is not allowed", method, consumer)
	}

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
	if c == nil {
		return status.Errorf(codes.Unauthenticated, "empty consumer")
	}

	consumer := c[0]
	method := info.FullMethod
	acl := srv.(*AdminManager).ACL

	valid := validate(acl, consumer, method)

	if !valid {
		return status.Errorf(codes.Unauthenticated, "method %s for consumer %s is not allowed", method, consumer)
	}

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
