package box

import (
	"github.com/suboat/go-contrib"

	"fmt"
	"math/rand"
	"runtime"
	"sort"
	"sync"
	"time"
)

// 容量测试中的并发统计
func (s *Scene) countConcurrency(l *sync.RWMutex, add int64, now, max *int64) {
	l.Lock()
	*now += add
	if *now > *max {
		*max = *now
	}
	l.Unlock()
	return
}

// 运行容量测试
func (s *Scene) run(form *FormScene, cache chan *ResultScene) (ret []*ResultScene, err error) {
	if err = form.Valid(); err != nil {
		return
	}
	if s.DefaultRobot == nil {
		// 未定义默认机器人
		if len(s.RobotArray) == 0 {
			return nil, contrib.ErrParamInvalid.SetVars("defaultRobot")
		}
		s.DefaultRobot = s.RobotArray[0]
	}
	if s.DefaultRobot.Scene == nil {
		s.DefaultRobot.Scene = s
	}

	//
	var (
		//
		category     = form.Category                            // 测试类型
		failBreak    = form.FailBreak                           // true: 遇错退出
		failFast     = form.FailFast                            // true: 遇到错误终止1个机器人
		failPerf     = PubFloatRound(float64(form.FailPerf), 4) // 性能下降阀值
		periodAction = form.GetPeriodAction()                   // 接口调用周期
		periodScene  = form.GetPeriodScene()                    // 场景时间跨度
		numInit      = form.NumInit                             // 每轮增加机器人数目
		numStep      = form.NumStep                             // 每轮增加机器人数目
		batchMax     = form.BatchMax                            // 最大运行轮数
		//
		data           []*ResultScene                   // 测试报告
		robots         [][]*Robot                       // 机器人运行结果
		batch          = 0                              // 目前运行第几轮
		batchRobot     = numInit                        // 本轮机器人数
		lastError      error                            // 最后一个错误
		concurrencyMax int64          = 0               // 累计最大并发
		concurrencyNow int64          = 0               // 当前并发
		countLock                     = &sync.RWMutex{} // 并发计数锁
	)

	// 运行
	fnRun := func() {
		// debug
		lastError = nil
		//s.Log.Debugf(`[scene-batch] #%d/%d robots:%d`, batch+1, batchMax, batchRobot)

		// 初始化机器人
		//s.wg.Add(batchRobot) // 机器人计数, 在执行前加好
		for len(s.RobotArray) < batchRobot {
			if _robot, _err := s.DefaultRobot.Copy(); _err != nil {
				err = _err
				return
			} else {
				_robot.Serial = len(s.RobotArray)
				_robot.Batch = batch
				s.RobotArray = append(s.RobotArray, _robot)
			}
		}

		// 遍历机器人
		for _, _d := range s.RobotArray {
			// 初始化结果槽
			robot := _d
			for len(robot.ResultArray) < len(robot.ActionArray) {
				robot.ResultArray = append(robot.ResultArray, nil)
			}

			// 动作计数
			robot.wg.Add(len(robot.ActionArray))
			go func() {
				// 机器人执行完动作后告知场景
				robot.wg.Wait()
				robot.Scene.wg.Done()
			}()

			// 动作执行
			go func() {
				defer PanicRecover(s.Log)

				// 在时间周期内随机起始时间
				if category == SceneCateCapacity || category == SceneCateStable {
					delay := time.Duration(int64(rand.Intn(int(periodAction)))) // 在场景时间内随机延时
					if delay > 0 {
						time.Sleep(delay)
					}
				}
				robot.TimeCreate = time.Now()

				// 顺序执行动作
				failNum := 0
				for _i, _d := range robot.ActionArray {
					idxAction := _i
					action := _d
					record := &RobotActionResult{
						Status: ActionStatusNormal,
					}

					// 运行或跳过
					if failNum > 0 && failFast {
						// 由于上一个动作错误将导致下一个错误
						record.Status = ActionStatusClose
						//record.Error = fmt.Errorf("fail fast")
					} else {
						// 运行前的参数准备
						if _err := action.RunBefore(robot, idxAction, batch); _err != nil {
							s.Log.Warnf(`[action-run-before] %s %d-%d `, robot.GetName(), batch, idxAction)
						}

						// 运行
						_start := time.Now()
						s.countConcurrency(countLock, 1, &concurrencyNow, &concurrencyMax)
						_ret, _err := action.Run(robot, idxAction, batch)
						s.countConcurrency(countLock, -1, &concurrencyNow, &concurrencyMax)
						_spent := time.Since(_start)
						// 结果
						record.Result = _ret
						record.Error = _err
						record.TimeCreate = _start
						record.TimeFinish = _start.Add(_spent)
						record.TimeSpent = _spent
						if record.Error != nil {
							record.Status = ActionStatusWarn
						}

						// 运行后的处理
						if _err := action.RunAfter(robot, idxAction, batch); _err != nil {
							s.Log.Warnf(`[action-run-after] %s %d-%d `, robot.GetName(), batch, idxAction)
						}

						// 统计耗时
						robot.TimeSpent += record.TimeSpent
					}

					// 错误计数
					if record.Status != ActionStatusNormal {
						//if record.Error != nil {
						failNum += 1
						if record.Error != nil {
							lastError = record.Error
						}
					}

					// 将动作结果放入列队
					robot.LockResult.Lock()
					robot.ResultArray[idxAction] = record
					robot.LockResult.Unlock()

					// next
					// 这个机器人完成了所有动作
					if _i == len(robot.ActionArray)-1 {
						robot.TimeFinish = time.Now()
						//robot.TimeSpent = robot.TimeFinish.Sub(robot.TimeCreate)
					}
					robot.wg.Done()
				}
			}()
		}
	}

	// log打印运行前参数
	s.Log.Infof(`[scene-run] params: %v`, PubJsonMust(form))

	// 执行与统计
	if s.FnBefore != nil {
		if err = s.FnBefore(s); err != nil {
			return
		}
	}
	for {
		// 执行一次测试
		var (
			report = &ResultScene{
				// 执行参数
				Category:     form.Category,
				FailBreak:    form.FailBreak,
				FailFast:     form.FailFast,
				FailPerf:     failPerf,
				BatchMax:     form.BatchMax,
				NumInit:      form.NumInit,
				NumStep:      form.NumStep,
				PeriodAction: form.PeriodAction,
				PeriodScene:  form.PeriodScene,
				// 本轮统计
				Batch:      batch + 1,  // 本轮测试是第几周期
				BatchRobot: batchRobot, // 本轮机器人数
				// 其它
				Params: form, // 执行参数
			}
		)

		// 上轮统计
		if len(data) > 0 {
			report.LastFailRate = data[len(data)-1].FailRate
			report.LastPerfAvg = data[len(data)-1].PerfTimeAvg
			report.LastPerf90Avg = data[len(data)-1].PerfTime90Avg
		}

		// 运行前准备
		if len(data) > 0 && periodScene > 0 {
			// 与上一轮期望的跨度有负差异
			if last := data[len(data)-1]; last.TimeEndLine.Unix() > 0 {
				now := time.Now()
				if diff := last.TimeEndLine.Sub(now); diff > 0 {
					// 比预期提前完成
					s.Log.Infof(`[scene-run-sleep] #%d %s <- %s sleep %.4fs after turn.`,
						last.Batch, PubTimeToStr(now), PubTimeToStr(last.TimeEndLine), diff.Seconds())
					time.Sleep(diff)
				} else {
					// 延迟完成
					s.Log.Warnf(`[scene-run-overlap] #%d %s <- %s overlap %.4fs`,
						last.Batch, PubTimeToStr(now), PubTimeToStr(last.TimeEndLine), -diff.Seconds())
				}
			}
		}

		// 运行测试
		report.TimeStart = time.Now()
		report.BatchText = fmt.Sprintf(`#%d. %s`, batch+1, report.TimeStart.Format("15:04:05"))
		if periodScene > 0 {
			report.TimeEndLine = report.TimeStart.Add(periodScene) // 期望结束的时间
		}
		s.Log.Infof(`[scene-run-%s] #%d/%d %du start %s`, report.Category, report.Batch, report.BatchMax,
			report.BatchRobot, PubTimeToStr(report.TimeStart))
		s.wg.Add(batchRobot) // 机器人计数
		go fnRun()
		//time.Sleep(time.Millisecond * 200) // 并发已发出 FIXME: 取更有说服力的sleep时间
		s.wg.Wait() // 等待所有机器人执行完

		// 运行测试: 将机器人与结果归档
		robots = append(robots, s.RobotArray)
		for i, robot := range s.RobotArray {
			if _err := robot.Close(); _err != nil {
				s.Log.Warnf(`[robot-close] %d-%d/%d"`, batch, i+1, batchRobot)
			}
		}
		s.RobotArray = []*Robot{}

		// 本轮统计: 耗时
		report.TimeEnd = time.Now()
		report.TimeRun = report.TimeEnd.Sub(report.TimeStart)

		// 本轮统计: 统计动作
		if batchRobot > 0 {
			var (
				total      = time.Duration(0) // 总耗时
				numFail    = 0                // 错误数目
				spentArray []int              // 总耗时
			)
			for _, d := range robots[len(robots)-1] {
				total += d.TimeSpent
				isSuccess := true
				// 机器人所有操作记录
				for _, r := range d.ResultArray {
					if r == nil || r.Status != ActionStatusNormal {
						numFail += 1
						isSuccess = false
						break
					}
					if r != nil && r.TimeSpent > 0 {
						if report.RespFastest == 0 {
							report.RespFastest = r.TimeSpent
						}
						if report.RespSlowest == 0 {
							report.RespSlowest = r.TimeSpent
						}
						if r.TimeSpent > report.RespSlowest {
							report.RespSlowest = r.TimeSpent
						}
						if r.TimeSpent < report.RespFastest {
							report.RespFastest = r.TimeSpent
						}
					}
				}
				if isSuccess && d.TimeSpent > 0 {
					spentArray = append(spentArray, int(d.TimeSpent))
				}
			}
			// 90%成功耗时
			sort.Ints(spentArray)
			from90 := int(float64(len(spentArray)) * 0.05)
			to90 := int(float64(len(spentArray)) * 0.95)
			if to90 < from90 {
				to90 = from90
			}
			total90 := 0
			for _, _d := range spentArray[from90:to90] {
				total90 += _d
			}
			if to90 > from90 {
				report.PerfTime90Avg = time.Duration(total90 / (to90 - from90))
			} else {
				report.PerfTime90Avg = time.Duration(total90)
			}
			// 统计耗时
			report.RespTotal = total
			report.PerfTimeAvg = total / time.Duration(batchRobot)
			// 错误率统计
			report.FailRate = PubFloatRound(float64(numFail)/float64(batchRobot), 4) // 4位小数
			// 性能下降率统计
			if len(data) > 0 {
				oldPerf := float64(data[len(data)-1].PerfTimeAvg)
				newPerf := float64(report.PerfTimeAvg)
				if oldPerf > 0 {
					report.PerfLossRate = PubFloatRound((newPerf-oldPerf)/oldPerf, 4)
				}
			}
		}

		// 本轮统计: 并发统计
		report.Concurrency = concurrencyMax
		concurrencyMax = 0
		concurrencyNow = 0

		// 本轮统计: 最后一个错误文本
		if lastError != nil {
			report.ErrText = lastError.Error() // 最后一个错误文本
		}

		// 本轮统计: TPS
		if report.RespSlowest > 0 {
			report.TpsMin = PubFloatRoundAuto(time.Second.Seconds() / report.RespSlowest.Seconds())
		}
		if report.RespFastest > 0 {
			report.TpsMax = PubFloatRoundAuto(time.Second.Seconds() / report.RespFastest.Seconds())
		}
		if report.PerfTimeAvg > 0 {
			report.TpsAvg = PubFloatRoundAuto(time.Second.Seconds() / report.PerfTimeAvg.Seconds())
		}
		if report.PerfTime90Avg > 0 {
			report.Tps90Avg = PubFloatRoundAuto(time.Second.Seconds() / report.PerfTime90Avg.Seconds())
		}

		// 累计统计: 累计耗时
		report.TotalTimeResp = report.RespTotal
		report.TotalTimeRun = report.TimeRun
		if len(data) > 0 {
			last := data[len(data)-1]
			report.TotalTimeRun += last.TotalTimeRun
			report.TotalTimeResp += last.TotalTimeResp
		}

		// 统计完成
		data = append(data, report)
		if len(robots) > 1 {
			robots[len(robots)-2] = nil // FIXME: 优化内存使用, 释放已统计的机器人
		}

		// 退出条件1: 出现了错误
		if lastError != nil {
			s.Log.Errorf(`[scene-break] #%d(this) failsRate:%.4f%% #%d(last) failsRate:%.4f%% lastErr: %v`,
				batch+1, report.FailRate*100, batch, report.LastFailRate*100, lastError)
			if report.Status == SceneStatusNormal && category == SceneCateCapacity && failBreak == true {
				// 容量测试遇错退出
				report.Status = SceneStatusFailBreak
			}
		}
		// 退出条件2: 性能下降
		if report.Status == SceneStatusNormal && failPerf > 0 && report.PerfLossRate >= failPerf {
			// 本轮性能下降超出阀值
			s.Log.Infof(`[scene-break] loss %f > %f, #%d(this) perf:%fS #%d(last) perf:%fs`,
				report.PerfLossRate, failPerf,
				batch+1, report.PerfTimeAvg.Seconds(), batch, report.LastPerfAvg.Seconds())
			if category == SceneCateCapacity {
				// 容量测试退出
				report.Status = SceneStatusFailPerf
			}
		}
		// 退出条件3: 到达边际
		if report.Status == SceneStatusNormal && batch >= batchMax-1 {
			report.Status = SceneStatusBatchMax
		}

		// 统计输出
		if cache != nil {
			cache <- report
		} else {
			s.Log.Infof(`[scene-batch] %s`, report.String())
		}

		// 退出
		if report.Status != SceneStatusNormal {
			break
		}

		// 进入下一轮
		lastError = nil
		batch += 1
		batchRobot += numStep
		runtime.GC() //
	}
	if s.FnAfter != nil {
		if err = s.FnAfter(s); err != nil {
			return
		}
	}

	// finish
	ret = data
	return
}
