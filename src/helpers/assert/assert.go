package assert

import (
	"errors"
	"fmt"
	"go-do-the-thing/src/helpers/slog"
	"reflect"
)

func NotEqual(actual any, notExpected any, source string, propName string) {
	if reflect.DeepEqual(actual, notExpected) {
		err := fmt.Errorf("%s should not be equal to %v", propName, notExpected)
		logger := slog.NewLogger(source)
		logger.Error(err, "")
		panic(err)
	}
}

func NoError(err error, source string, msg string, params ...any) {
	if err == nil {
		return
	}
	logger := slog.NewLogger(source)
	logger.Error(err, msg, params...)
	panic(err)
}

func IsTrue(isTrue bool, source string, msg string, params ...any) {
	if isTrue {
		return
	}
	err := errors.New("unexpected situation")
	logger := slog.NewLogger(source)
	logger.Error(err, msg, params...)
	panic(err)
}

func NotNil(val any, source string, msg string) {
	if !isNil(val) {
		return
	}
	err := errors.New("should not be nil")
	logger := slog.NewLogger(source)
	logger.Error(err, msg)
	panic(err)
}

func isNil(i any) bool {
	if i == nil {
		return true
	}

	value := reflect.ValueOf(i)
	kind := value.Kind()

	switch kind {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return value.IsNil()
	default:
		return false
	}
}
