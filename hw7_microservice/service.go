package main

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/tap"
	"net"

	"google.golang.org/grpc"
)

type BizManager struct {
}

type AdminManager struct {
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
		grpc.InTapHandle(statistics),
		grpc.UnaryInterceptor(aclChecker),
	)
	RegisterBizServer(server, &BizManager{})
	RegisterAdminServer(server, &AdminManager{})
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

func statistics(ctx context.Context, info *tap.Info) (context.Context, error) {
	return ctx, nil
}

func aclChecker(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	fmt.Println(md)
	reply, err := handler(ctx, req)
	return reply, err
}
