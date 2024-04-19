package orch

import (
	"context"
	"fmt"
)

type ErrorDapr struct {
	Ctx context.Context
	Msg string
	Err error
}

func (e ErrorDapr) Error() string {
	return fmt.Sprintf("%s: %s", e.Msg, e.Err)
}

func (e ErrorDapr) Unwrap() error {
	return e.Err
}

type ErrorApp struct {
	Ctx context.Context
	Err error
}

func (e ErrorApp) Error() string {
	return e.Err.Error()
}

func (e ErrorApp) Unwrap() error {
	return e.Err
}
