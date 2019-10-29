package box

import (
	"time"
)

// 一个执行函数 step: 执行位置0起始, batch: 执行批次,第几次执行
type ActionOneFn func(u *Robot, step, batch int, act *ActionOne) (ret interface{}, err error)

// 一个动作实例
type ActionOne struct {
	Name     string      // 动作命名
	Fn       ActionOneFn // 执行函数
	FnBefore ActionOneFn // 执行函数
	FnAfter  ActionOneFn // 执行函数
	Interval time.Time   // 执行间隔
}

// 动作印记
type ActionOnePrint struct {
	Status     int       // 动作状态
	TimeCreate time.Time // 开始时间
	TimeFinish time.Time // 完成时间
	//
	Action *ActionOne // 父级动作
}

//
func (d *ActionOne) Run(u *Robot, step, batch int) (ret interface{}, err error) {
	if d.Fn != nil {
		return d.Fn(u, step, batch, d)
	} else {
		u.Scene.Log.Warnf(`[action-fn] "%s" undefined in %d-%d`, d.Name, step, batch)
	}
	return
}

//
func (d *ActionOne) RunBefore(u *Robot, step, batch int) (err error) {
	if d.FnBefore != nil {
		_, err = d.FnBefore(u, step, batch, d)
	}
	return
}

//
func (d *ActionOne) RunAfter(u *Robot, step, batch int) (err error) {
	if d.FnBefore != nil {
		_, err = d.FnAfter(u, step, batch, d)
	}
	return
}

//
func (d *ActionOne) GetName() (ret string) {
	return d.Name
}
