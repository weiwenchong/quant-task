package logic

import (
	"context"
	. "github.com/wenchong-wei/quant-task/pub"
	"github.com/wenchong-wei/quant-task/service/task"
	"log"
)

type GrpcTask struct {
}

func (m *GrpcTask) CreatePriceTask(ctx context.Context, req *CreatePriceTaskReq) (*CreatePriceTaskRes, error) {
	fun := "GrpcTask.CreatePriceTask -->"
	log.Printf("%s incall", fun)
	err := task.FactoryPriceTask().CreatePriceTask(ctx, req)
	if err != nil {
		return nil, err
	}

	log.Printf("%s succeed req:%v", fun, req)
	return &CreatePriceTaskRes{}, nil
}
