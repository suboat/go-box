package box

import (
	"github.com/suboat/go-contrib"

	"time"
)

// 测试参数
type FormScene struct {
	// 测试类型
	Category string // capacity|surge|stable
	// 测试参数
	FailBreak    bool    // true:容量测试时,接口无法正确返回时即退出测试 false:接口错误仍继续执行到期望的最大周期。【默认true】
	FailFast     bool    // 仅对测试单元执行接口数大于1时起作用。true:遇错停止执行本单元的余下接口调用 false:遇错仍继续执行本单元余下接口调用，不管余下调用是否成功，本单元均标记为失败。【默认true】
	FailPerf     float32 // 设置数值大于0有效。如设为0.7，代表容量测试第N轮的TPS比N-1轮TPS少70%及以上，则退出容量测试。【默认0】
	BatchMax     int     // 容量测试最大执行轮数理论值，实际执行会受FailBreak、FailPerf参数影响。【默认100】
	NumInit      int     // 容量测试初始机器人数。【默认100】
	NumStep      int     // 容量测试机器人递增数。【默认100】
	PeriodAction int64   // 容量测试的"接口调用周期"参数，单位毫秒。【默认4000】
	PeriodScene  int64   // 容量测试中一个场景(Scene)的时间跨度理论值，实际执行会受场景中调用最慢的一次接口影响，单位毫秒。【默认10000】
}

// 容量测试
type FormCapacity struct {
	FailBreak    bool    //
	FailFast     bool    //
	FailPerf     float32 //
	BatchMax     int     //
	NumInit      int     //
	NumStep      int     //
	PeriodAction int64   //
	PeriodScene  int64   //
}

// 浪涌测试
type FormSurge struct {
	BatchMax    int   //
	NumInit     int   //
	PeriodScene int64 //
}

// 稳定性测试
type FormStable struct {
	NumInit      int   //
	PeriodAction int64 //
	PeriodScene  int64 //
	//
	Duration int // 持续时间,单位秒
}

//
func (d *FormScene) Valid() (err error) {
	if d == nil {
		return contrib.ErrParamUndefined
	}
	if d.NumInit <= 0 {
		return contrib.ErrParamInvalid.SetVars("numInit")
	}
	return
}

// 取接口执行周期
func (d *FormScene) GetPeriodAction() time.Duration {
	return time.Duration(d.PeriodAction) * time.Millisecond
}

// 取每轮理论时间跨度
func (d *FormScene) GetPeriodScene() time.Duration {
	return time.Duration(d.PeriodScene) * time.Millisecond
}

//
func (d *FormCapacity) Valid() (err error) {
	return
}

//
func (d *FormCapacity) GetForm() (ret *FormScene, err error) {
	if err = d.Valid(); err != nil {
		return
	}
	ret = new(FormScene)
	ret.Category = SceneCateCapacity
	ret.FailBreak = d.FailBreak
	ret.FailFast = d.FailFast
	ret.FailPerf = d.FailPerf
	ret.BatchMax = d.BatchMax
	ret.NumInit = d.NumInit
	ret.NumStep = d.NumStep
	ret.PeriodAction = d.PeriodAction
	ret.PeriodScene = d.PeriodScene
	return
}

//
func (d *FormSurge) Valid() (err error) {
	if d == nil {
		return contrib.ErrParamUndefined
	}
	return
}

//
func (d *FormSurge) GetForm() (ret *FormScene, err error) {
	if err = d.Valid(); err != nil {
		return
	}
	ret = new(FormScene)
	ret.Category = SceneCateSurge
	ret.FailBreak = false
	ret.FailFast = true
	ret.FailPerf = 0
	ret.BatchMax = d.BatchMax
	ret.NumInit = d.NumInit
	ret.NumStep = 0
	ret.PeriodAction = 0
	ret.PeriodScene = d.PeriodScene
	return
}

//
func (d *FormStable) Valid() (err error) {
	if d == nil {
		return contrib.ErrParamUndefined
	}
	return
}

//
func (d *FormStable) GetForm() (ret *FormScene, err error) {
	if err = d.Valid(); err != nil {
		return
	}
	ret = new(FormScene)
	ret.Category = SceneCateStable
	ret.FailBreak = false
	ret.FailFast = true
	ret.FailPerf = 0
	ret.BatchMax = int((time.Duration(d.Duration) * time.Second) / (time.Duration(d.PeriodScene) * time.Millisecond))
	ret.NumInit = d.NumInit
	ret.NumStep = 0
	ret.PeriodAction = d.PeriodAction
	ret.PeriodScene = d.PeriodScene
	return
}
