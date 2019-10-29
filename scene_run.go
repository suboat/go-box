package box

import ()

// 执行容量测试
func (s *Scene) RunCapacity(form *FormCapacity, cache chan *ResultScene) (ret []*ResultScene, err error) {
	var formScene *FormScene
	if formScene, err = form.GetForm(); err != nil {
		return
	}
	return s.run(formScene, cache)
}

// 执行浪涌测试
func (s *Scene) RunSurge(form *FormSurge, cache chan *ResultScene) (ret []*ResultScene, err error) {
	var formScene *FormScene
	if formScene, err = form.GetForm(); err != nil {
		return
	}
	return s.run(formScene, cache)
}

// 执行稳定性测试
func (s *Scene) RunStable(form *FormStable, cache chan *ResultScene) (ret []*ResultScene, err error) {
	var formScene *FormScene
	if formScene, err = form.GetForm(); err != nil {
		return
	}
	return s.run(formScene, cache)
}
