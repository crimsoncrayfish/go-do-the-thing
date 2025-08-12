package assert

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"
)

func NotEqual(actual any, notExpected any, source string, propName string) {
	if reflect.DeepEqual(actual, notExpected) {
		err := fmt.Errorf("%s should not be equal to %v", propName, notExpected)
		slog.Error("assertion failed", "error", err, "source", source)
		panic(err)
	}
}

func NoError(err error, source string, msg string, params ...any) {
	if err == nil {
		return
	}
	slog.Error(msg, "error", err, "source", source)
	panic(err)
}

func IsTrue(isTrue bool, source string, msg string, params ...any) {
	if isTrue {
		return
	}
	err := errors.New("unexpected situation")
	slog.Error(msg, "error", err, "source", source)
	panic(err)
}

func NotNil(val any, source string, msg string) {
	if !isNil(val) {
		return
	}
	err := errors.New("should not be nil")
	slog.Error(msg, "error", err, "source", source)
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
