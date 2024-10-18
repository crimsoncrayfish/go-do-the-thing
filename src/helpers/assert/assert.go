package assert

import (
	"errors"
	"go-do-the-thing/src/helpers/slog"
)

func NoError(err error, logger slog.Logger, msg string, params ...any) {
	if err == nil {
		return
	}
	logger.Error(err, msg, params...)
	panic(err)
}

func IsTrue(isTrue bool, msg string, params ...any) {
	if isTrue {
		return
	}
	err := errors.New("unexpected situation")
	logger := slog.NewLogger("Panic at the disco")
	logger.Error(err, msg, params...)
	panic(err)
}
