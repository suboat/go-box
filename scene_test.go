package box

import (
	"github.com/stretchr/testify/require"

	"testing"
	"time"
)

func testScene(t testing.TB) (ret *Scene) {
	scene := NewScene(nil)
	return scene
}

//
func TestMain(m *testing.M) {
	m.Run()
}

// 测试场景容量
func Test_SceneCapacity(t *testing.T) {
	as := require.New(t)
	scene := testScene(t)

	// 定义1个机器
	robot := NewRobot(&Robot{
		Name: "robot",
		FnClose: func(d *Robot) (err error) {
			d.Scene.Log.Debugf("[robot-close] %s", d.GetName())
			return
		},
	})

	// 定义机器人的动作
	action1 := NewActionOne(&ActionOne{
		Name: "action1",
		Fn: func(u *Robot, step, batch int, act *ActionOne) (ret interface{}, err error) {
			u.Scene.Log.Debugf(`[action1] %s %s-%d-%d`, u.GetName(), act.Name, batch, step)
			time.Sleep(time.Millisecond * 200)
			return
		},
	})
	action2 := NewActionOne(&ActionOne{
		Name: "action2",
		Fn: func(u *Robot, step, batch int, act *ActionOne) (ret interface{}, err error) {
			u.Scene.Log.Debugf(`[action2] %s %s-%d-%d`, u.GetName(), act.Name, batch, step)
			time.Sleep(time.Millisecond * 200)
			return
		},
	})
	as.Nil(robot.AddAction(action1))
	as.Nil(robot.AddAction(action2))

	// 初始化机器人
	scene.DefaultRobot = robot

	// 容量测试
	scene.Log.SetLevel(4)
	ret, err := scene.RunCapacity(&FormCapacity{
		NumInit: 500,
	}, nil)
	as.Nil(err)
	t.Log(ret)
}

// chan
func Test_Chan(t *testing.T) {
	c := make(chan int)
	go func() {
		time.Sleep(time.Second * 2)
		c <- 1
		close(c)
		println("222")
	}()
	//
	select {
	case d := <-c:
		println("asdad", d)
		break
	case <-time.After(time.Second):
		break
	}
	t.Log("haha")
	time.Sleep(time.Second * 2000)
}
