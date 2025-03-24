package assert

import (
	"errors"
	"go-do-the-thing/src/helpers/slog"
)

type Source struct {
	Name string
}

func NoError(err error, source Source, msg string, params ...any) {
	if err == nil {
		return
	}
	logger := slog.NewLogger(source.Name)
	logger.Error(err, msg, params...)
	panic(err)
}

func IsTrue(isTrue bool, source Source, msg string, params ...any) {
	if isTrue {
		return
	}
	err := errors.New("unexpected situation")
	logger := slog.NewLogger(source.Name)
	logger.Error(err, msg, params...)
	panic(err)
}
