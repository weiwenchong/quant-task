package task

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/gogf/gf/os/gcron"
	. "github.com/wenchong-wei/quant-task/pub"
	"github.com/wenchong-wei/quant-task/service/model/cache"
	"github.com/wenchong-wei/quant-task/service/model/dao"
	"github.com/wenchong-wei/quant-task/service/util"
	"log"
	"strconv"
	"time"
)

const (
	// redis 任务key
	TIME_TASK          = "time_task"
	PRICE_TASK_GREATER = "price_task_%s_greater"
	PRICE_TASK_LESS    = "price_task_%s_less"
	ASSETS             = "assets"
)

func InitTask() {
	// 时间任务，每秒执行
	gcron.Add("* * * * * *", func() {
		now := time.Now().Unix()
		go CheckTimeTask(now)
		go UpdateAndCheckPriceTask()
	})
}

type priceTask struct {
	Infos []*dao.TaskInfo
}

func FactoryPriceTask() *priceTask {
	return &priceTask{}
}

func FactoryPriceTaskByInfos(infos []*dao.TaskInfo) *priceTask {
	return &priceTask{infos}
}

func CheckTimeTask(now int64) {
	vals, err := cache.Client.ZRangeByScore(TIME_TASK, redis.ZRangeBy{
		Min:    "-1",
		Max:    strconv.FormatInt(now, 10),
		Offset: 0,
		Count:  0,
	}).Result()
	if err != nil {
		log.Printf("CheckTimeTask ZRangeByScore err:%v time:%d", err, now)
		return
	}
	_, err = cache.Client.ZRemRangeByScore(TIME_TASK, "-1", strconv.FormatInt(now, 10)).Result()
	if err != nil {
		log.Printf("CheckTimeTask ZRemRangeByScore err:%v time:%d", err, now)
	}
	ids := make([]int64, 0)
	for _, v := range vals {
		id, _ := strconv.ParseInt(v, 10, 64)
		ids = append(ids, id)
	}
	_ = FactoryPriceTask().doTimeTask(context.TODO(), ids)
	log.Printf("CheckTimeTask success, time:%d", now)
}

func (m *priceTask) CreatePriceTask(ctx context.Context, req *CreatePriceTaskReq) error {
	fun := "pirceTask.CreatePriceTask -->"
	if len(req.Tasks) == 0 {
		return errors.New("empty tasks")
	}

	memMap := make(map[string]bool)
	members := make([]interface{}, 0)
	insertMap := make([]map[string]interface{}, 0)
	now := time.Now().Unix()
	for _, task := range req.Tasks {
		insertMap = append(insertMap, map[string]interface{}{
			"uid":           task.Uid,
			"sourceservice": req.Source,
			"assettype":     task.AssetType,
			"assetcode":     task.AssetCode,
			"cond":          task.Condition,
			"price":         task.Price,
			"starttime":     task.StartTime,
			"message":       task.Message,
			"status":        0,
			"ct":            now,
		})
		memMap[util.GetAssetCode(task.AssetType, task.AssetCode)] = true
	}

	for k, _ := range memMap {
		members = append(members, k)
	}
	if len(members) > 0 {
		_, err := cache.Client.SAdd(ASSETS, members...).Result()
		if err != nil {
			log.Printf("%s SAdd err:%v", fun, err)
		}
	}

	id, err := dao.Insert(ctx, dao.DB, dao.TASK_INFO, insertMap)
	if err != nil {
		log.Printf("%s Insert task err:%v", fun, err)
		return err
	}
	ids := make([]int64, 0)
	for i := 0; i < len(insertMap); i++ {
		ids = append(ids, id+int64(i))
	}

	err = dao.SelectList(ctx, dao.DB, dao.TASK_INFO, map[string]interface{}{"id in": ids}, &m.Infos)
	if err != nil {
		log.Printf("%s SelectList Task err:%v", fun, err)
		return err
	}
	timePriceInfos := make([]*dao.TaskInfo, 0)
	priceInfos := make([]*dao.TaskInfo, 0)
	for _, info := range m.Infos {
		if info.StartTime == 0 {
			priceInfos = append(priceInfos, info)
		} else {
			timePriceInfos = append(timePriceInfos, info)
		}
	}

	if len(timePriceInfos) > 0 {
		err = FactoryPriceTaskByInfos(timePriceInfos).startTimePriceTask(ctx)
		if err != nil {
			log.Printf("%s startTimePriceTask err:%v", fun, err)
		}
	}
	if len(priceInfos) > 0 {
		err = FactoryPriceTaskByInfos(priceInfos).startPriceTask(ctx)
		if err != nil {
			log.Printf("%s startPriceTask err:%v", fun, err)
		}
	}
	return nil
}

func (m *priceTask) startTimePriceTask(ctx context.Context) error {
	fun := "priceTask.CreateTimePriceTask -->"
	members := make([]redis.Z, 0)
	for _, info := range m.Infos {
		members = append(members, redis.Z{
			Score:  float64(info.StartTime),
			Member: info.Id,
		})
	}
	_, err := cache.Client.ZAdd(TIME_TASK, members...).Result()
	if err != nil {
		log.Printf("%s Zadd err:%v", fun, err)
		return err
	}
	return nil
}

func (m *priceTask) startPriceTask(ctx context.Context) error {
	fun := "priceTask.startPriceTask -->"
	for _, info := range m.Infos {
		tmp := ""
		switch info.Cond {
		case int32(PriceCondition_LESS):
			tmp = PRICE_TASK_LESS
		case int32(PriceCondition_GREATER):
			tmp = PRICE_TASK_GREATER
		default:
			log.Printf("%s invalid info.Cond:%d", fun, info.Cond)
			//return errors.New("invalid info.Cond")
		}
		_, err := cache.Client.ZAdd(fmt.Sprintf(fmt.Sprintf(tmp, util.GetAssetCode(info.AssetType, info.AssetCode))), redis.Z{
			Score:  float64(info.Price),
			Member: info.Id,
		}).Result()
		if err != nil {
			log.Printf("%s Zadd err:%v", fun, err)
		}
	}
	return nil
}

func (m *priceTask) doTimeTask(ctx context.Context, ids []int64) error {
	fun := "priceTask.doTimeTask -->"
	err := dao.SelectList(ctx, dao.DB, dao.TASK_INFO, map[string]interface{}{"id in": ids}, &m.Infos)
	if err != nil {
		log.Printf("%s SelectList err:%v", fun, err)
		return err
	}
	timeInfos := make([]*dao.TaskInfo, 0)
	priceInfos := make([]*dao.TaskInfo, 0)
	for _, info := range m.Infos {
		if info.AssetType != 0 {
			priceInfos = append(priceInfos, info)
		} else {
			timeInfos = append(timeInfos, info)
		}
	}
	_ = FactoryPriceTaskByInfos(priceInfos).startPriceTask(ctx)
	_ = FactoryPriceTaskByInfos(timeInfos).doTask(ctx)
	return nil
}

func (m *priceTask) doTask(ctx context.Context) error {
	fun := "priceTask.getTaskids -->"
	now := time.Now().Unix()
	for _, info := range m.Infos {
		err := cache.Publish(fmt.Sprintf("task_%d", info.SourceService), info.Message)
		if err != nil {
			log.Printf("%s Publish err:%v", fun, err)
			continue
		}
		_, err = dao.Update(ctx, dao.DB, dao.TASK_INFO, map[string]interface{}{"id": info.Id}, map[string]interface{}{"status": 1, "ut": now})
		if err != nil {
			log.Printf("%s Update err:%v", fun, err)
			continue
		}
	}
	return nil
}
