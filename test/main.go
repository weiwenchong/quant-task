package main

import (
	"context"
	"fmt"
	"github.com/wenchong-wei/quant-task/Adapter"
	"github.com/wenchong-wei/quant-task/pub"
	"github.com/wenchong-wei/quant-task/service/util"
	"time"
)

var (
	// 上证 深证交易时间段
	shzhStart1 int64 = 9*3600 + 30*60
	shzhEnd1   int64 = 11*3600 + 30*60
	shzhStart2 int64 = 13 * 3600
	shzhEnd2   int64 = 15 * 3600

	// 港股交易时间段
	hkStart1 int64 = 9*3600 + 30*60
	hkEnd1   int64 = 12 * 3600
	hkStart2 int64 = 13 * 3600
	hkEnd2   int64 = 16 * 3600

	// 美股交易时间段
	gbEnd   int64 = 8 * 3600
	gbStart int64 = 16 * 3600
)

func main() {
	now := time.Now().Unix()
	week := time.Now().Weekday()
	todayTime := now - util.DayBeginStamp(now)
	if week >= 0 && week <= 4 && ((todayTime >= shzhStart1 && todayTime <= shzhEnd1) || (todayTime >= shzhStart2 && todayTime <= shzhEnd2)) {
		fmt.Println(1)
	}
	if week >= 0 && week <= 4 && ((todayTime >= hkStart1 && todayTime <= hkEnd1) || (todayTime >= hkStart2 && todayTime <= hkEnd2)) {
		fmt.Println(2)
	}

	// 美股任务
	if (week == 0 && todayTime >= gbStart) || (week == 5 && todayTime <= gbEnd) || (week >= 1 && week <= 4 && (todayTime <= gbEnd || todayTime >= gbStart)) {
		fmt.Println(3)
	}

	Adapter.InitClient()
	fmt.Println(Adapter.Client.CreatePriceTask(context.TODO(), &pub.CreatePriceTaskReq{
		Source: 1,
		Tasks: []*pub.PriceTask{{
			AssetType: 3,
			AssetCode: "BABA",
			Condition: pub.PriceCondition_GREATER,
			Price:     165000,
			StartTime: 1629806700,
			Message:   "1",
			Uid:       1,
		}},
	}))

}
