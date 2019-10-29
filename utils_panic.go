package box

import (
	"github.com/suboat/go-contrib"
	"github.com/suboat/go-contrib/log"

	"fmt"
	"runtime"
	"strings"
)

var (
	// LogPanic 输出panic错误
	LogPanic Logger
)

// PanicRecover 统一处理panic
func PanicRecover(logger Logger) {
	r := recover()
	if r == nil {
		return
	}
	if logger == nil && LogPanic != nil {
		logger = LogPanic
	} else {
		logger = log.Log
	}
	logger.Errorf(`[panic-recover] "%s" %v`, panicIdentify(), r)
}

// PanicRecoverError 统一处理panic, 并更新error
func PanicRecoverError(logger Logger, err *error) {
	r := recover()
	if r == nil {
		return
	}
	if logger == nil && LogPanic != nil {
		logger = LogPanic
	} else {
		logger = log.Log
	}
	logger.Errorf(`[panic-recover] "%s" %v`, panicIdentify(), r)
	if err != nil {
		*err = contrib.ErrPanicRecover.SetVars(r)
	}
	return
}

// 定位panic位置 参考自: https://gist.github.com/swdunlop/9629168
func panicIdentify() string {
	var (
		pc [16]uintptr
		n  = runtime.Callers(3, pc[:])
	)

	for _, pc := range pc[:n] {
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		_fnName := fn.Name()
		if strings.HasPrefix(_fnName, "runtime.") {
			continue
		}
		file, line := fn.FileLine(pc)

		//
		var (
			_fnNameDir = strings.Split(_fnName, "/")
			_fnNameLis = strings.Split(_fnName, ".")
			_fnNameSrc string
		)
		if len(_fnNameDir) > 1 {
			_fnNameSrc = _fnNameDir[0] + "/" + _fnNameDir[1] + "/"
		} else {
			_fnNameSrc = _fnNameDir[0]
		}
		fnName := _fnNameLis[len(_fnNameLis)-1]

		// file
		_pcLis := strings.Split(file, _fnNameSrc)
		filePath := strings.Join(_pcLis[1:], "")

		return fmt.Sprintf("%s:%d|%s", filePath, line, fnName)
	}

	return "unknown"
}
