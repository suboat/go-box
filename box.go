package box

import (
	"sync"
	"time"
)

// Logger 日志输出
type Logger interface {
	SetLevel(level uint32)
	GetLevel() uint32
	//
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Debugln(args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Infoln(args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Warnln(args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Errorln(args ...interface{})
	Print(args ...interface{})
	Printf(format string, args ...interface{})
	Println(args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatalln(args ...interface{})
	Panic(args ...interface{})
	Panicf(format string, args ...interface{})
	Panicln(args ...interface{})
}

// 版本信息
type Version struct {
	Major int // 主版本号:向前不兼容	变化通常意味着模块的巨大的变化
	Minor int // 次版本号:对同级兼容	通常只反映了一些较大的更改, 比如模块的API增加等
	Patch int // 补丁版本:对同级兼容	通常情况下如果只是对模块的修改而不影响API接口
	// optional
	Model  string     // 模块名称
	Hash   string     // 可执行文件哈希
	Commit *string    // 代码提交哈希
	Build  *time.Time // 模块编译时间
}

// 配置文件
type Config struct {
	sync.RWMutex
	pathYaml   *string                 // yaml配置文件保存地址
	savePoint  interface{}             //
	hookChange func(interface{}) error //
	comments   map[string]string       // 配置文件的备注信息
	silent     bool                    // true: 不打印日志
}
