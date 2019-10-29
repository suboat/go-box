package box

import (
	"fmt"
)

// 机器人关闭
type RobotClose func(d *Robot) (err error)

// 复制用户
func (d *Robot) Copy() (ret *Robot, err error) {
	var (
		data *Robot
	)
	data = NewRobot(nil)
	data.Name = d.Name
	data.Serial = d.Serial + 1
	data.Batch = d.Batch
	data.ActionWindow = d.ActionWindow
	data.ActionArray = []Action{}
	for _, d := range d.ActionArray {
		data.ActionArray = append(data.ActionArray, d)
	}
	data.ResultArray = []*RobotActionResult{}
	data.Scene = d.Scene
	data.FnClose = d.FnClose
	data.IsCopy = true

	//
	ret = data
	return
}

// 机器人称呼
func (d *Robot) GetName() (ret string) {
	return fmt.Sprintf("%s-%d-%d", d.Name, d.Batch, d.Serial)
}

// 添加行为
func (d *Robot) AddAction(a Action) (err error) {
	d.ActionArray = append(d.ActionArray, a)
	return
}

// 关闭
func (d *Robot) Close() (err error) {
	if d.FnClose != nil {
		return d.FnClose(d)
	}
	return
}
