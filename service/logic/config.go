package logic

import "github.com/weiwenchong/quant-task/service/task"

func InitLogic() {
	// 初始化调用rpc

	// 服务自己的init
	task.InitTask()
}
