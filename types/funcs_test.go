package types

import (
	"context"
	"testing"
)

func TestIsFuncType(t *testing.T) {
	if !IsFuncType(func() {}) {
		t.Errorf("expected func() to be a valid FuncType")
	}
	if !IsFuncType(func() error { return nil }) {
		t.Errorf("expected func() error to be a valid FuncType")
	}
	if !IsFuncType(func(context.Context) {}) {
		t.Errorf("expected func(context.Context) to be a valid FuncType")
	}
	if !IsFuncType(func(context.Context) error { return nil }) {
		t.Errorf("expected func(context.Context) error to be a valid FuncType")
	}

	// Test the Invalid case
	if IsFuncType(func(int) error { return nil }) {
		t.Errorf("expected func(int) error to not be a valid FuncType")
	}
}

func TestFuncTypeWrap(t *testing.T) {
	if FuncTypeWrap(func(i int) {}) != nil {
		t.Errorf("expected func(int) to return nil")
	}

	if FuncTypeWrap(func() {}) == nil {
		t.Errorf("expected func() to return a function")
	}

	if FuncTypeWrap(func() error { return nil }) == nil {
		t.Errorf("expected func() error to return a function")
	}

	if FuncTypeWrap(func(context.Context) {}) == nil {
		t.Errorf("expected func(context.Context) to return a function")
	}

	if FuncTypeWrap(func(context.Context) error { return nil }) == nil {
		t.Errorf("expected func(context.Context) error to return a function")
	}
}
