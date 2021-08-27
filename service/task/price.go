package task

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/wenchong-wei/quant-task/service/model/cache"
	"github.com/wenchong-wei/quant-task/service/model/dao"
	"github.com/wenchong-wei/quant-task/service/util"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	// 上证 深证交易时间段
	shzhStart1 int64 = 9*3600 + 29*60
	shzhEnd1   int64 = 11*3600 + 31*60
	shzhStart2 int64 = 13*3600 - 60
	shzhEnd2   int64 = 16*3600 + 60

	// 港股交易时间段
	hkStart1 int64 = 9*3600 + 30*60 - 60
	hkEnd1   int64 = 12*3600 + 60
	hkStart2 int64 = 13*3600 - 60
	hkEnd2   int64 = 16*3600 + 60

	// 美股交易时间段
	gbEnd   int64 = 5*3600 - 60
	gbStart int64 = 21*3600 + 30*60 + 60
)

func UpdateAndCheckPriceTask() {
	fun := "UpdateAndCheckPriceTask -->"
	as, err := cache.Client.SMembers(ASSETS).Result()
	if err != nil {
		log.Printf("%s SMembers err:%v", fun, err)
		return
	}

	shzh := []string{}
	hk := []string{}
	gb := []string{}
	needUpdateAssets := []string{}
	ids := []int64{}

	for _, a := range as {
		if len(a) != 0 {
			if a[0] == '1' || a[0] == '2' {
				shzh = append(shzh, a)
			} else if a[0] == '4' {
				hk = append(hk, a)
			} else if a[0] == '3' {
				gb = append(gb, a)
			}
		}
	}

	now := time.Now().Unix()
	week := time.Now().Weekday()
	todayTime := now - util.DayBeginStamp(now)
	fmt.Println()
	// 上证深证任务
	if week >= 1 && week <= 5 && ((todayTime >= shzhStart1 && todayTime <= shzhEnd1) || (todayTime >= shzhStart2 && todayTime <= shzhEnd2)) {
		needUpdateAssets = append(needUpdateAssets, shzh...)
	}

	// 港股任务
	if week >= 1 && week <= 5 && ((todayTime >= hkStart1 && todayTime <= hkEnd1) || (todayTime >= hkStart2 && todayTime <= hkEnd2)) {
		needUpdateAssets = append(needUpdateAssets, hk...)
	}

	// 美股任务
	if (week == 1 && todayTime >= gbStart) || (week == 6 && todayTime <= gbEnd) || (week >= 2 && week <= 5 && (todayTime <= gbEnd || todayTime >= gbStart)) {
		needUpdateAssets = append(needUpdateAssets, gb...)
	}

	// 更新价格
	FactoryUpdatePrices(needUpdateAssets).updatePrices()

	for _, a := range needUpdateAssets {
		err, is := FactoryPrice(a).getTaskids()
		if err != nil {
			log.Printf("%s getTaskids asset:%s err:%v", fun, a, err)
			continue
		}
		ids = append(ids, is...)
	}

	infos := make([]*dao.TaskInfo, 0)
	ctx := context.TODO()
	err = dao.SelectList(ctx, dao.DB, dao.TASK_INFO, map[string]interface{}{"id in": ids}, &infos)
	if err != nil {
		log.Printf("%s SelectList err:%v", fun, err)
		return
	}
	// 执行价格任务
	_ = FactoryPriceTaskByInfos(infos).doTask(ctx)

	log.Printf("%s succeed assets:%v", fun, as)
}

type price struct {
	Asset string
}

func FactoryPrice(asset string) *price {
	return &price{asset}
}

func (m *price) getTaskids() (error, []int64) {
	fun := "price.getTaskids -->"
	ss := []string{}
	p, err := cache.Client.Get(m.Asset).Result()
	if err != nil {
		log.Printf("%s Get Price err:%v", fun, err)
		return err, nil
	}
	vals, err := cache.Client.ZRangeByScore(fmt.Sprintf(PRICE_TASK_GREATER, m.Asset), redis.ZRangeBy{
		Min:    "-1",
		Max:    p,
		Offset: 0,
		Count:  0,
	}).Result()
	if err != nil {
		log.Printf("%s ZRangeByScore greater asset:%s err:%v", fun, m.Asset, err)
	} else {
		_, err = cache.Client.ZRemRangeByScore(fmt.Sprintf(PRICE_TASK_GREATER, m.Asset), "-1", p).Result()
		if err != nil {
			log.Printf("%s ZRemRangeByScore greater err:%v asset:%s", fun, err, m.Asset)
		}
	}
	ss = append(ss, vals...)
	vals1, err := cache.Client.ZRangeByScore(fmt.Sprintf(PRICE_TASK_LESS, m.Asset), redis.ZRangeBy{
		Min:    p,
		Max:    "9223372036854775807",
		Offset: 0,
		Count:  0,
	}).Result()
	if err != nil {
		log.Printf("%s ZRangeByScore less asset:%s err:%v", fun, m.Asset, err)
	} else {
		_, err = cache.Client.ZRemRangeByScore(fmt.Sprintf(PRICE_TASK_LESS, m.Asset), p, "9223372036854775807").Result()
		if err != nil {
			log.Printf("%s ZRemRangeByScore less err:%v asset:%s", fun, err, m.Asset)
		}
	}
	ss = append(ss, vals1...)
	ids := make([]int64, 0)
	for _, s := range ss {
		id, _ := strconv.ParseInt(s, 10, 64)
		ids = append(ids, id)
	}
	return nil, ids
}

type updatePrices struct {
	Assets          []string
	ReqNameAssetMap map[string]string
	AssetInfoMap    map[string]*assetInfo
}

type assetInfo struct {
	Symbol  string
	Current float64
}

func FactoryUpdatePrices(assets []string) *updatePrices {
	return &updatePrices{Assets: assets}
}

func (m *updatePrices) updatePrices() {
	fun := "updatePrices.updatePrices -->"
	m.getAssetReqName()
	m.getAssetInfo()

	actual := []string{}
	for k, v := range m.AssetInfoMap {
		_, err := cache.Client.Set(k, strconv.FormatInt(int64(v.Current*1000), 10), 0).Result()
		if err != nil {
			log.Printf("%s Set err:%v", fun, err)
			continue
		}
		actual = append(actual, k)
	}
	log.Printf("%s succeed, asset:%v actual:%v", fun, m.Assets, actual)
}

func (m *updatePrices) getAssetReqName() {
	fun := "updatePrices.getAssetReqName -->"
	m.ReqNameAssetMap = make(map[string]string)
	for _, asset := range m.Assets {
		reqName, err := util.GetAssetHttpReqName(asset)
		if err != nil {
			log.Printf("%s GetAssetHttpReqName err:%v", fun, err)
			continue
		}
		m.ReqNameAssetMap[reqName] = asset
	}
}

type SnowBallBody struct {
	Infos []*assetInfo `json:"data"`
}

func (m *updatePrices) getAssetInfo() {
	fun := "updatePrices.getAssetInfo -->"
	m.AssetInfoMap = make(map[string]*assetInfo)
	reqNames := []string{}
	for k, _ := range m.ReqNameAssetMap {
		reqNames = append(reqNames, k)
	}
	reqs := strings.Join(reqNames, ",")

	res, err := http.Get("https://stock.xueqiu.com/v5/stock/realtime/quotec.json?symbol=" + reqs)
	if err != nil {
		log.Printf("%s http.Get err:%v", fun, err)
		return
	}
	bs, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("%s ioutil.ReadAll err:%v", fun, err)
		return
	}

	body := SnowBallBody{}
	err = json.Unmarshal(bs, &body)
	if err != nil {
		log.Printf("%s json.Unmarshal err:%v", fun, err)
		return
	}

	for _, info := range body.Infos {
		m.AssetInfoMap[m.ReqNameAssetMap[info.Symbol]] = info
	}
}
