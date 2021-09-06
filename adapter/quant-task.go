package adapter

import (
	"github.com/gogf/gf/os/gcron"
	"github.com/wenchong-wei/quant-task/pub"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"log"
)

const PORT = "172.17.0.4:10001"

var (
	Client pub.TaskClient
	Conn   *grpc.ClientConn
)

func InitClient() {
	log.Printf("initClient start")
	gcron.Add("* * * * * *", func() {
		var err error
		if Conn == nil || Conn.GetState() == connectivity.TransientFailure || Conn.GetState() == connectivity.Shutdown {
			Conn, err = grpc.Dial(PORT, grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				log.Printf("InitClient quant-order err:%v", err)
			}
			Client = pub.NewTaskClient(Conn)
		}
	})
	log.Printf("initClient end")
}
