package types

import (
	"context"
)

// FuncType indicates a prototype of build job function
type FuncType int

// FuncTypes
const (
	InvalidType FuncType = iota
	VoidType
	ErrorType
	ContextVoidType
	ContextErrorType
)

// FuncVoid is the simplist job function
type FuncVoid func()

// FuncError may return an error
type FuncError func() error

// FuncContext takes a context for the job
type FuncContext func(context.Context)

// FuncContextError takes a context and may return an error
type FuncContextError func(context.Context) error

// IsFuncType tests if a function is one of FuncType
func IsFuncType(fn interface{}) bool {
	switch fn.(type) {
	case func():
		return true
	case func() error:
		return true
	case func(context.Context):
		return true
	case func(context.Context) error:
		return true
	}
	return false
}

// FuncTypeWrap wraps a valid FuncType to FuncContextError
func FuncTypeWrap(fn interface{}) FuncContextError {
	if IsFuncType(fn) {
		switch f := fn.(type) {
		case func():
			return func(context.Context) error {
				f()
				return nil
			}
		case func() error:
			return func(context.Context) error {
				return f()
			}
		case func(context.Context):
			return func(ctx context.Context) error {
				f(ctx)
				return nil
			}
		case func(context.Context) error:
			return f
		}
	}
	return nil
}
