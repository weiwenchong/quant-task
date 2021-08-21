package logic

import (
	"context"
	. "github.com/wenchong-wei/quant-task/pub"
	"log"
)

type GrpcTask struct {
}

func (m *GrpcTask) CreatePriceTask(ctx context.Context, req *CreatePriceTaskReq) (*CreatePriceTaskRes, error) {
	fun := "GrpcTask.CreatePriceTask -->"
	log.Printf("%s incall", fun)
}
