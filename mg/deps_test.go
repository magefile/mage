package mg

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestDepsRunOnce(t *testing.T) {
	done := make(chan struct{})
	f := func() {
		done <- struct{}{}
	}
	go Deps(f, f)
	select {
	case <-done:
		// cool
	case <-time.After(time.Millisecond * 100):
		t.Fatal("func not run in a reasonable amount of time.")
	}
	select {
	case <-done:
		t.Fatal("func run twice!")
	case <-time.After(time.Millisecond * 100):
		// cool... this should be plenty of time for the goroutine to have run
	}
}

func TestDepsOfDeps(t *testing.T) {
	ch := make(chan string, 3)
	// this->f->g->h
	h := func() {
		ch <- "h"
	}
	g := func() {
		Deps(h)
		ch <- "g"
	}
	f := func() {
		Deps(g)
		ch <- "f"
	}
	Deps(f)

	res := <-ch + <-ch + <-ch

	if res != "hgf" {
		t.Fatal("expected h then g then f to run, but got " + res)
	}
}

func TestSerialDeps(t *testing.T) {
	ch := make(chan string, 3)
	// this->f->g->h
	h := func() {
		ch <- "h"
	}
	g := func() {
		ch <- "g"
	}
	f := func() {
		SerialDeps(g, h)
		ch <- "f"
	}
	Deps(f)

	res := <-ch + <-ch + <-ch

	if res != "ghf" {
		t.Fatal("expected g then h then f to run, but got " + res)
	}
}

func TestDepError(t *testing.T) {
	// TODO: this test is ugly and relies on implementation details. It should
	// be recreated as a full-stack test.

	f := func() error {
		return errors.New("ouch")
	}
	defer func() {
		err := recover()
		if err == nil {
			t.Fatal("expected panic, but didn't get one")
		}
		actual := fmt.Sprint(err)
		if actual != "ouch" {
			t.Fatalf(`expected to get "ouch" but got "%s"`, actual)
		}
	}()
	Deps(f)
}

func TestDepFatal(t *testing.T) {
	f := func() error {
		return Fatal(99, "ouch!")
	}
	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("expected panic, but didn't get one")
		}
		actual := fmt.Sprint(v)
		if actual != "ouch!" {
			t.Fatalf(`expected to get "ouch!" but got "%s"`, actual)
		}
		err, ok := v.(error)
		if !ok {
			t.Fatalf("expected recovered val to be error but was %T", v)
		}
		code := ExitStatus(err)
		if code != 99 {
			t.Fatalf("Expected exit status 99, but got %v", code)
		}
	}()
	Deps(f)
}

func TestDepTwoFatal(t *testing.T) {
	f := func() error {
		return Fatal(99, "ouch!")
	}
	g := func() error {
		return Fatal(11, "bang!")
	}
	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("expected panic, but didn't get one")
		}
		actual := fmt.Sprint(v)
		// order is non-deterministic, so check for both orders
		if actual != "ouch!\nbang!" && actual != "bang!\nouch!" {
			t.Fatalf(`expected to get "ouch!" and "bang!" but got "%s"`, actual)
		}
		err, ok := v.(error)
		if !ok {
			t.Fatalf("expected recovered val to be error but was %T", v)
		}
		code := ExitStatus(err)
		// two different error codes returns, so we give up and just use error
		// code 1.
		if code != 1 {
			t.Fatalf("Expected exit status 1, but got %v", code)
		}
	}()
	Deps(f, g)
}

func TestDepWithUnhandledFunc(t *testing.T) {
	defer func() {
		err := recover()
		_, ok := err.(error)
		if !ok {
			t.Fatal("Expected type error from panic")
		}
	}()
	var notValid = func(a string) string {
		return a
	}
	Deps(notValid)
}

func TestDepsErrors(t *testing.T) {
	var hRan, gRan, fRan int64

	h := func() error {
		atomic.AddInt64(&hRan, 1)
		return errors.New("oops")
	}
	g := func() {
		Deps(h)
		atomic.AddInt64(&gRan, 1)
	}
	f := func() {
		Deps(g, h)
		atomic.AddInt64(&fRan, 1)
	}

	defer func() {
		err := recover()
		if err == nil {
			t.Fatal("expected f to panic")
		}
		if hRan != 1 {
			t.Fatalf("expected h to run once, but got %v", hRan)
		}
		if gRan > 0 {
			t.Fatalf("expected g to panic before incrementing gRan to run, but got %v", gRan)
		}
		if fRan > 0 {
			t.Fatalf("expected f to panic before incrementing fRan to run, but got %v", fRan)
		}
	}()
	f()
}

func TestDepPanicNonError(t *testing.T) {
	f := func() {
		panic("string panic value")
	}
	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("expected panic, but didn't get one")
		}
		actual := fmt.Sprint(v)
		if actual != "string panic value" {
			t.Fatalf(`expected "string panic value" but got %q`, actual)
		}
		err, ok := v.(error)
		if !ok {
			t.Fatalf("expected recovered val to be error but was %T", v)
		}
		code := ExitStatus(err)
		if code != 1 {
			t.Fatalf("Expected exit status 1 for non-error panic, but got %v", code)
		}
	}()
	Deps(f)
}

func TestCtxDepsPassesContext(t *testing.T) {
	type ctxKey string
	var got context.Context
	f := func(ctx context.Context) {
		got = ctx
	}
	ctx := context.WithValue(context.Background(), ctxKey("key"), "value")
	CtxDeps(ctx, f)
	if got == nil {
		t.Fatal("context was not passed to dependency")
	}
	if got.Value(ctxKey("key")) != "value" {
		t.Fatal("context value was not propagated")
	}
}

func TestChangeExit(t *testing.T) {
	tests := []struct {
		name string
		old  int
		nw   int
		want int
	}{
		{"both zero", 0, 0, 0},
		{"old zero new nonzero", 0, 5, 5},
		{"old nonzero new zero", 5, 0, 5},
		{"same nonzero", 5, 5, 5},
		{"different nonzero", 5, 3, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := changeExit(tt.old, tt.nw)
			if got != tt.want {
				t.Errorf("changeExit(%d, %d) = %d, want %d", tt.old, tt.nw, got, tt.want)
			}
		})
	}
}

func TestSerialCtxDeps(t *testing.T) {
	type ctxKey string
	ch := make(chan string, 2)
	f := func(ctx context.Context) {
		if ctx.Value(ctxKey("key")) != "value" {
			panic("missing context value")
		}
		ch <- "f"
	}
	g := func(ctx context.Context) {
		ch <- "g"
	}
	ctx := context.WithValue(context.Background(), ctxKey("key"), "value")
	SerialCtxDeps(ctx, f, g)

	first := <-ch
	second := <-ch
	if first != "f" || second != "g" {
		t.Fatalf("expected serial execution f then g, got %s then %s", first, second)
	}
}
