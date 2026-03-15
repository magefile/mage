package mg

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

func TestFuncCheck(t *testing.T) {
	hasContext, isNamespace, err := checkF(func() {}, nil)
	if err != nil {
		t.Error(err)
	}
	if hasContext {
		t.Error("func does not have context")
	}
	if isNamespace {
		t.Error("func is not on a namespace")
	}
	hasContext, isNamespace, err = checkF(func() error { return nil }, nil)
	if err != nil {
		t.Error(err)
	}
	if hasContext {
		t.Error("func does not have context")
	}
	if isNamespace {
		t.Error("func is not on a namespace")
	}
	hasContext, isNamespace, err = checkF(func(context.Context) {}, nil)
	if err != nil {
		t.Error(err)
	}
	if !hasContext {
		t.Error("func has context")
	}
	if isNamespace {
		t.Error("func is not on a namespace")
	}
	hasContext, isNamespace, err = checkF(func(context.Context) error { return nil }, nil)
	if err != nil {
		t.Error(err)
	}
	if !hasContext {
		t.Error("func has context")
	}
	if isNamespace {
		t.Error("func is not on a namespace")
	}

	hasContext, isNamespace, err = checkF(Foo.Bare, nil)
	if err != nil {
		t.Error(err)
	}
	if hasContext {
		t.Error("func does not have context")
	}
	if !isNamespace {
		t.Error("func is on a namespace")
	}

	hasContext, isNamespace, err = checkF(Foo.Error, nil)
	if err != nil {
		t.Error(err)
	}
	if hasContext {
		t.Error("func does not have context")
	}
	if !isNamespace {
		t.Error("func is on a namespace")
	}

	hasContext, isNamespace, err = checkF(Foo.BareCtx, nil)
	if err != nil {
		t.Error(err)
	}
	if !hasContext {
		t.Error("func has context")
	}
	if !isNamespace {
		t.Error("func is  on a namespace")
	}
	hasContext, isNamespace, err = checkF(Foo.CtxError, nil)
	if err != nil {
		t.Error(err)
	}
	if !hasContext {
		t.Error("func has context")
	}
	if !isNamespace {
		t.Error("func is  on a namespace")
	}

	hasContext, isNamespace, err = checkF(Foo.CtxErrorArgs, []interface{}{1, "s", true, time.Second})
	if err != nil {
		t.Error(err)
	}
	if !hasContext {
		t.Error("func has context")
	}
	if !isNamespace {
		t.Error("func is on a namespace")
	}

	hasContext, isNamespace, err = checkF(func(int, bool, string, time.Duration) {}, []interface{}{1, true, "s", time.Second})
	if err != nil {
		t.Error(err)
	}
	if hasContext {
		t.Error("func does not have context")
	}
	if isNamespace {
		t.Error("func is not on a namespace")
	}

	// Test an Invalid case
	_, _, err = checkF(func(*int) error { return nil }, nil)
	if err == nil {
		t.Error("expected func(*int) error to be invalid")
	}

	defer func() {
		if r := recover(); r != nil {
			t.Error("expected a nil function argument to be handled gracefully")
		}
	}()
	_, _, err = checkF(nil, []interface{}{1, 2})
	if err == nil {
		t.Error("expected a nil function argument to be invalid")
	}
}

func TestF(t *testing.T) {
	var (
		ctxOut interface{}
		iOut   int
		sOut   string
		bOut   bool
		dOut   time.Duration
	)
	f := func(cctx context.Context, ii int, ss string, bb bool, dd time.Duration) error {
		ctxOut = cctx
		iOut = ii
		sOut = ss
		bOut = bb
		dOut = dd
		return nil
	}

	ctx := context.Background()
	i := 1776
	s := "abc124"
	b := true
	d := time.Second

	CtxDeps(ctx, F(f, i, s, b, d))
	if ctxOut != ctx {
		t.Error(ctxOut)
	}
	if iOut != i {
		t.Error(iOut)
	}
	if bOut != b {
		t.Error(bOut)
	}
	if dOut != d {
		t.Error(dOut)
	}
	if sOut != s {
		t.Error(sOut)
	}
}

func TestFTwice(t *testing.T) {
	var called int64
	f := func(int) {
		atomic.AddInt64(&called, 1)
	}

	Deps(F(f, 5), F(f, 5), F(f, 1))
	if called != 2 {
		t.Fatalf("Expected to be called 2 times, but was called %d", called)
	}
}

func ExampleF() {
	f := func(i int) {
		fmt.Println(i)
	}

	// we use SerialDeps here to ensure consistent output, but this works with all Deps functions.
	SerialDeps(F(f, 5), F(f, 1))
	// output:
	// 5
	// 1
}

func TestFNamespace(t *testing.T) {
	ctx := context.Background()
	i := 1776
	s := "abc124"
	b := true
	d := time.Second

	fn := F(Foo.CtxErrorArgs, i, s, b, d)
	err := fn.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFNilError(t *testing.T) {
	fn := F(func() error { return nil })
	err := fn.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestFVariadic(t *testing.T) {
	fn := F(func(args ...string) {
		if !reflect.DeepEqual(args, []string{"a", "b"}) {
			t.Errorf("Wrong args, got %v, want [a b]", args)
		}
	}, "a", "b")
	err := fn.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	fn = F(func(_ string, _ ...string) {}, "a", "b1", "b2")
	err = fn.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	fn = F(func(_ ...string) {})
	err = fn.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	func() {
		defer func() {
			err, ok := recover().(error)
			if !ok || err == nil || err.Error() != "too few arguments for target, got 0 for func(string, ...string)" {
				t.Fatal(err)
			}
		}()
		F(func(_ string, _ ...string) {})
	}()
}

type Foo Namespace

func (Foo) Bare() {}

func (Foo) Error() error { return nil }

func (Foo) BareCtx(context.Context) {}

func (Foo) CtxError(context.Context) error { return nil }

func (Foo) CtxErrorArgs(_ context.Context, _ int, _ string, _ bool, _ time.Duration) error {
	return nil
}

func TestFNonFunction(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for non-function")
		}
	}()
	F("not a function")
}

func TestFTooManyArgs(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for too many args")
		}
	}()
	F(func(int) {}, 1, 2)
}

func TestFWrongArgType(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for wrong arg type")
		}
	}()
	F(func(int) {}, "not an int")
}

func TestFUnsupportedArgType(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for unsupported arg type")
		}
	}()
	F(func(*int) {}, (*int)(nil))
}

func TestFTooManyReturns(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for too many return values")
		}
	}()
	F(func() (int, error) { return 0, nil })
}

func TestFNonErrorReturn(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for non-error return")
		}
	}()
	F(func() int { return 0 })
}

func TestFReturnsError(t *testing.T) {
	origErr := errors.New("boom")
	fn := F(func() error { return origErr })
	err := fn.Run(context.Background())
	if !errors.Is(err, origErr) {
		t.Fatalf("expected 'boom' error, got %v", err)
	}
}

func TestFnID(t *testing.T) {
	fn1 := F(func(int) {}, 1)
	fn2 := F(func(int) {}, 2)
	fn3 := F(func(int) {}, 1)
	if fn1.ID() == fn2.ID() {
		t.Error("different args should produce different IDs")
	}
	if fn1.ID() != fn3.ID() {
		t.Error("same args should produce same IDs")
	}
}

func TestFnName(t *testing.T) {
	fn := F(func() {})
	if fn.Name() == "" {
		t.Error("expected non-empty function name")
	}
}
