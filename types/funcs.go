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
