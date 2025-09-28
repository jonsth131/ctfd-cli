package tui

import (
	"context"
	"errors"
	"fmt"
)

func createErrMsg(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return errMsg{fmt.Errorf("Request timed out")}
	} else if errors.Is(err, context.Canceled) {
		return errMsg{fmt.Errorf("Request was canceled")}
	} else {
		return errMsg{err}
	}
}
