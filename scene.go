package box

import (
	"github.com/suboat/go-contrib/log"

	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// 测试类型
const (
	SceneCateCapacity = "capacity" // 容量测试
	SceneCateSurge    = "surge"    // 浪涌测试
	SceneCateStable   = "stable"   // 稳定测试
)

// 场景状态
const (
	SceneStatusNormal    int = iota // 0: 正常
	SceneStatusFailBreak            // 1: 出现了错误
	SceneStatusFailPerf             // 2: 性能下降超出预期
	SceneStatusBatchMax             // 3: 执行到了最大周期
)

// 动作状态
const (
	ActionStatusNormal   int = iota // 0: 已完成
	ActionStatusWarn                // 1: 异常:返回错误
	ActionStatusFreeze              // 2: 暂停:执行超时
	ActionStatusCreating            // 3: 正在执行
	ActionStatusClose               // 4: 关闭:主动停止执行
	ActionStatusDelete              // 5: 已删除;已经标记为删除,等待系统逐步回收资源后彻底删除
)

// 默认参数
var (
	DefaultScenePeriodWindow      = time.Second * 5  // 场景默认五秒一个周期
	DefaultSceneNumCpu            = runtime.NumCPU() // 并发数默认等于cpu数
	DefaultSceneCapacityMaxLoop   = 1000             // 容量测试最大周期
	DefaultSceneCapacityRobotStep = 5                // 容量测试每期递增人数
	DefaultSceneCapacityBreakRate = 0.8              // 容量测试退出的衰减阀值
)

// 一个场景
type Scene struct {
	Name       string   // 场景名
	RobotArray []*Robot // 场景中的用户
	FnBefore   SceneFn  // 场景执行前的准备函数,如载入登录用户数据
	FnAfter    SceneFn  // 场景执行后的收尾函数,如程序执行被中断时的结果保存
	//
	NumCpu int // 程序并发数
	//
	Log          Logger // 日志
	DefaultRobot *Robot // 默认机器人
	//
	wg sync.WaitGroup // 机器人并行后集合
}
type SceneFn func(s *Scene) (err error)

// 一个用户
type Robot struct {
	Name         string               // 用户名
	Batch        int                  // 批次
	Serial       int                  // 编号
	ActionWindow time.Time            // 动作执行区间,在多少时间内把动作做完
	ActionArray  []Action             // 要做的动作
	ResultArray  []*RobotActionResult // 动作执行结果
	FnClose      RobotClose           // 关闭机器人
	//
	TimeCreate time.Time     // 开始时间
	TimeFinish time.Time     // 完成时间
	TimeSpent  time.Duration // 耗时
	//
	Scene  *Scene // 父级场景
	IsCopy bool   // true: 是复制而来
	//
	Context    context.Context //
	LockResult sync.RWMutex    //
	wg         sync.WaitGroup  //
}
type RobotActionResult struct {
	Result     interface{}   // 返回结果
	Error      error         // 返回
	Status     int           // 动作状态
	TimeCreate time.Time     // 开始时间
	TimeFinish time.Time     // 完成时间
	TimeSpent  time.Duration // 耗时
}

// 一个动作
type Action interface {
	Run(u *Robot, step, batch int) (result interface{}, err error) // 执行 step: 执行位置0起始, batch: 执行批次,第几次执行
	RunBefore(u *Robot, step, batch int) error                     // 执行前
	RunAfter(u *Robot, step, batch int) error                      // 执行后
	// get
	GetName() string // 动作名称
}

// 测试结果
type ResultScene struct {
	// 执行参数: 各字段含义见FormScene
	Category     string  `json:"category"`     //
	FailBreak    bool    `json:"failBreak"`    //
	FailFast     bool    `json:"failFast"`     //
	FailPerf     float64 `json:"failPerf"`     //
	BatchMax     int     `json:"batchMax"`     //
	NumInit      int     `json:"numInit"`      //
	NumStep      int     `json:"numStep"`      //
	PeriodAction int64   `json:"periodAction"` //
	PeriodScene  int64   `json:"periodScene"`  //
	// 上轮统计
	LastFailRate  float64       `json:"lastFailRate"`  // 上一轮容量测试的错误率
	LastPerfAvg   time.Duration `json:"lastPerfAvg"`   // 上一轮平均耗时
	LastPerf90Avg time.Duration `json:"lastPerf90Avg"` // 上一轮90%平均耗时
	// 本轮统计
	Status        int           `json:"status"`        // 本轮测试状态
	Batch         int           `json:"batch"`         // 本轮测试是第几周期
	BatchRobot    int           `json:"batchRobot"`    // 本轮机器人数
	BatchText     string        `json:"batchText"`     // 本轮名称
	TimeStart     time.Time     `json:"timeStart"`     // 本轮开始时间
	TimeEnd       time.Time     `json:"timeEnd"`       // 本轮结束时间
	TimeEndLine   time.Time     `json:"timeEndLine"`   // 本轮期望结束时间
	TimeRun       time.Duration `json:"timeRun"`       // 本轮运行时间
	Concurrency   int64         `json:"concurrency"`   // 本轮最大并发
	ErrText       string        `json:"errText"`       // 最后一个错误文本
	FailRate      float64       `json:"failRate"`      // 本轮错误率
	TpsMax        float64       `json:"tpsMax"`        // 高峰TPS
	TpsMin        float64       `json:"tpsMin"`        // 谷底TPS
	TpsAvg        float64       `json:"tpsAvg"`        // 平均TPS
	Tps90Avg      float64       `json:"tps90Avg"`      // 90%平均TPS
	PerfTimeAvg   time.Duration `json:"perfTimeAvg"`   // 请求平均耗时
	PerfTime90Avg time.Duration `json:"perfTime90Avg"` // 90%请求耗时
	PerfTime90Std time.Duration `json:"perfTime90Std"` // 90%请求的标准差
	PerfLossRate  float64       `json:"perfLossRate"`  // 本轮性能下降率
	RespTotal     time.Duration `json:"perfTimeTotal"` // 响应总耗时
	RespFastest   time.Duration `json:"respFastest"`   // 响应最快请求
	RespSlowest   time.Duration `json:"respSlowest"`   // 响应最慢请求
	// 累计统计
	TotalTimeRun  time.Duration `json:"totalTime"`     // 累计运行时间
	TotalTimeResp time.Duration `json:"totalTimeResp"` // 累计响应时间
	// 其它
	Params      *FormScene              `json:"-"` // 执行参数
	ActionArray []*ResultCapacityAction `json:"-"` // 动作执行结果
}
type ResultCapacityAction struct {
}

// 结果摘要
func (d *ResultScene) String() (ret string) {
	ret = fmt.Sprintf(`#%d/%d-%s %du conc:%d loss:%f tps:%f over:%fs err:%s`,
		d.Batch, d.BatchMax, d.Category, d.BatchRobot, d.Concurrency, d.PerfLossRate, d.Tps90Avg,
		d.TimeEnd.Sub(d.TimeEndLine).Seconds(),
		d.ErrText)
	return
}

// 创建新场景
func NewScene(s *Scene) (d *Scene) {
	if s != nil {
		d = s
	} else {
		d = new(Scene)
	}
	if d.NumCpu == 0 {
		d.NumCpu = DefaultSceneNumCpu
	}
	if d.Log == nil {
		d.Log = log.Log
	}
	return
}

// 创建新用户
func NewRobot(s *Robot) (d *Robot) {
	if s != nil {
		d = s
	} else {
		d = new(Robot)
	}
	return
}

// 创建动作
func NewActionOne(s *ActionOne) (d *ActionOne) {
	if s != nil {
		d = s
	} else {
		d = new(ActionOne)
	}
	return
}
