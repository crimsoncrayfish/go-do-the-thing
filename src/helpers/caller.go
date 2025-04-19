package helpers

import "runtime"

func CallerName() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return ""
	}
	fname := runtime.FuncForPC(pc)
	if fname == nil {
		return ""
	}
	return fname.Name()
}
