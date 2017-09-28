package mg_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/pkg/errors"
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
	// TODO: this test is ugly and relies on implementation details. It should
	// be recreated as a full-stack test.

	f := func() error {
		return mg.Fatal(99, "ouch!")
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
