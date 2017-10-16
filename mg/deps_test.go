package mg_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/magefile/mage/mg"
)

func TestDepsRunOnce(t *testing.T) {
	done := make(chan struct{})
	f := func() {
		done <- struct{}{}
	}
	go mg.Deps(f, f)
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
		mg.Deps(h)
		ch <- "g"
	}
	f := func() {
		mg.Deps(g)
		ch <- "f"
	}
	mg.Deps(f)

	res := <-ch + <-ch + <-ch

	if res != "hgf" {
		t.Fatal("expected h then g then f to run, but got " + res)
	}
}

func TestDepError(t *testing.T) {
	// TODO: this test is ugly and relies on implementation details. It should
	// be recreated as a full-stack test.

	f := func() error {
		return errors.New("ouch!")
	}
	defer func() {
		err := recover()
		if err == nil {
			t.Fatal("expected panic, but didn't get one")
		}
		actual := fmt.Sprint(err)
		if "ouch!" != actual {
			t.Fatalf(`expected to get "ouch!" but got "%s"`, actual)
		}
	}()
	mg.Deps(f)
}

func TestDepFatal(t *testing.T) {
	f := func() error {
		return mg.Fatal(99, "ouch!")
	}
	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("expected panic, but didn't get one")
		}
		actual := fmt.Sprint(v)
		if "ouch!" != actual {
			t.Fatalf(`expected to get "ouch!" but got "%s"`, actual)
		}
		err, ok := v.(error)
		if !ok {
			t.Fatalf("expected recovered val to be error but was %T", v)
		}
		code := mg.ExitStatus(err)
		if code != 99 {
			t.Fatalf("Expected exit status 99, but got %v", code)
		}
	}()
	mg.Deps(f)
}

func TestDepTwoFatal(t *testing.T) {
	f := func() error {
		return mg.Fatal(99, "ouch!")
	}
	g := func() error {
		return mg.Fatal(11, "bang!")
	}
	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("expected panic, but didn't get one")
		}
		actual := fmt.Sprint(v)
		// order is non-deterministic, so check for both orders
		if "ouch!\nbang!" != actual && "bang!\nouch!" != actual {
			t.Fatalf(`expected to get "ouch!" and "bang!" but got "%s"`, actual)
		}
		err, ok := v.(error)
		if !ok {
			t.Fatalf("expected recovered val to be error but was %T", v)
		}
		code := mg.ExitStatus(err)
		// two different error codes returns, so we give up and just use error
		// code 1.
		if code != 1 {
			t.Fatalf("Expected exit status 1, but got %v", code)
		}
	}()
	mg.Deps(f, g)
}

func TestDepWithUnhandledFunc(t *testing.T) {
	defer func() {
		err := recover()
		expected := "Invalid type for dependent function: func(string) string. Dependencies must be func(), func() error, func(context.Context) or func(context.Context) error"
		actual, ok := err.(error)
		if !ok {
			t.Fatalf("Expected type string from panic")
		}
		if actual.Error() != expected {
			t.Fatalf("Expected panic %v but got %v", expected, err)
		}
	}()
	var NotValid func(string) string = func(a string) string {
		return a
	}
	mg.Deps(NotValid)
}
