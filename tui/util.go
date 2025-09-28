package tui

import (
	"context"
	"errors"

	"github.com/jonsth131/ctfd-cli/tui/constants"
)

type (
	errMsg struct{ err error }
)

func (e errMsg) Error() string { return e.err.Error() }

func createErrMsg(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return errMsg{errors.New("Request timed out")}
	} else if errors.Is(err, context.Canceled) {
		return errMsg{errors.New("Request was canceled")}
	} else {
		return errMsg{err}
	}
}

func renderError(err error) string {
	if err == nil {
		return ""
	}
	return constants.ErrStyle(err.Error())
}
