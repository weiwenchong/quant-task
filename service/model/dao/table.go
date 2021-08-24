package dao

const (
	TASK_INFO = "task_task_info"
)

type TaskInfo struct {
	Id            int64  `json:"id" ddb:"id"`
	Uid           int64  `json:"uid" ddb:"uid"`
	SourceService int32  `json:"sourceservice" ddb:"sourceservice"`
	AssetType     int32  `json:"assettype" ddb:"assettype"`
	AssetCode     string `json:"assetcode" ddb:"assetcode"`
	Cond          int32  `json:"cond" ddb:"cond"`
	Price         int64  `json:"price" ddb:"price"`
	StartTime     int64  `json:"starttime" ddb:"starttime"`
	Message       string `json:"message" ddb:"message"`
	Status        int32  `json:"status" ddb:"status"`
	Ct            int64  `json:"ct" ddb:"ct"`
	Ut            int64  `json:"ut" ddb:"ut"`
}
