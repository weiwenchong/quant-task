package main

import (
	. "github.com/wenchong-wei/quant-task/pub"
	"github.com/wenchong-wei/quant-task/service/logic"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	log.Printf("service start")
	lis, err := net.Listen("tcp", PORT)
	if err != nil {
		log.Panicf("quant-task listen err:%v", err)
		return
	}

	s := grpc.NewServer()
	RegisterTaskServer(s, new(logic.GrpcTask))
	log.Printf("task start")
	if err = s.Serve(lis); err != nil {
		log.Panicf("quant-task serve err:%v", err)
	}
}