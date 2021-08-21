package task

import (
	"github.com/gogf/gf/os/gcron"
)

const (
	TIME_TASK = "time_task"
	PRICE_TASK = "price_task"
)

func InitTimeTask() {
	gcron.Add("* * * * * *", func(){

	})
}

type timeTask struct {}

func FactoryTimeTaskBy() *timeTask {

}

func (m *timeTask) AddTask