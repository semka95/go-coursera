package main

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
)

type BizManager struct {
}

type AdminManager struct {
}

func (srv *BizManager) Check(ctx context.Context, in *Nothing) (*Nothing, error) {
	return &Nothing{Dummy:true}, nil
}
func (srv *BizManager) Add(ctx context.Context, in *Nothing) (*Nothing, error) {
	return &Nothing{Dummy:true}, nil
}
func (srv *BizManager) Test(ctx context.Context, in *Nothing) (*Nothing, error) {
	return &Nothing{Dummy:true}, nil
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

	server := grpc.NewServer()

	RegisterBizServer(server, &BizManager{})
	RegisterAdminServer(server, &AdminManager{})
	fmt.Println("starting server at ", listenAddr)
	server.Serve(lis)
	return nil
}
