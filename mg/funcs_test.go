package mg

import (
	"context"
	"testing"
)

func TestWrapFunc(t *testing.T) {
	// we can ignore errors here, since the error is always the same and the
	// FuncType will be nil if there's an error.
	d, _ := wrapFn(func() {})
	if _, ok := d.(voidFn); !ok {
		t.Errorf("expected func() to be a VoidFn, but was %v", d)
	}
	d, _ = wrapFn(func() error { return nil })
	if _, ok := d.(errorFn); !ok {
		t.Errorf("expected func() error to be a errorFn, but was %v", d)
	}
	d, _ = wrapFn(func(context.Context) {})
	if _, ok := d.(contextVoidFn); !ok {
		t.Errorf("expected func(context.Context) to be a contextVoidFn, but was %v", d)
	}
	d, _ = wrapFn(func(context.Context) error { return nil })
	if _, ok := d.(contextErrorFn); !ok {
		t.Errorf("expected func(context.Context) error to be a contextErrorFn but was %v", d)
	}

	d, _ = wrapFn(Foo.Bare)
	if _, ok := d.(namespaceVoidFn); !ok {
		t.Errorf("expected Foo.Bare to be a namespaceVoidFn but was %v", d)
	}

	d, _ = wrapFn(Foo.Error)
	if _, ok := d.(namespaceErrorFn); !ok {
		t.Errorf("expected Foo.Error to be a namespaceErrorFn but was %v", d)
	}
	d, _ = wrapFn(Foo.BareCtx)
	if _, ok := d.(namespaceContextVoidFn); !ok {
		t.Errorf("expected Foo.BareCtx to be a namespaceContextVoidFn but was %v", d)
	}
	d, _ = wrapFn(Foo.CtxError)
	if _, ok := d.(namespaceContextErrorFn); !ok {
		t.Errorf("expected Foo.CtxError to be a namespaceContextErrorFn but was %v", d)
	}

	// Test the Invalid case
	d, err := wrapFn(func(int) error { return nil })
	if d != nil {
		t.Errorf("expected func(int) error to be invalid, but was %v", d)
	}
	if err == nil {
		t.Error("expected func(int) error to not be a valid FuncType, but got nil error.")
	}
}

type Foo Namespace

func (Foo) Bare() {}

func (Foo) Error() error { return nil }

func (Foo) BareCtx(context.Context) {}

func (Foo) CtxError(context.Context) error { return nil }
