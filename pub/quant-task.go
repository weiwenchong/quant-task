package pub

import (
	"google.golang.org/grpc"
	"log"
)

const PORT = ":10002"

var Client TaskClient

func InitClient() {
	conn, err := grpc.Dial(PORT, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Panicf("InitClient quant-order err:%v", err)
	}
	Client = NewTaskClient(conn)
	go func() {
		defer func() {
			log.Printf("conn close start")
			conn.Close()
			log.Printf("conn close")
		}()
		select{}
	}()

}