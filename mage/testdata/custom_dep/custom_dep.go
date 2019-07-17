//+build mage

package main

import (
	"context"
	"fmt"

	"github.com/magefile/mage/mg"
)

type ParameterizedDep struct {
	i int
}

func (pd ParameterizedDep) Identify() string {
	return fmt.Sprintf("main.ParameterizedDep(%d)", pd.i)
}

func (pd ParameterizedDep) Run(ctx context.Context) error {
	fmt.Printf("%d\n", pd.i)
	return nil
}

func Main() {
	mg.Deps(
		ParameterizedDep{1},
		ParameterizedDep{2},
		ParameterizedDep{3},
		ParameterizedDep{4},
		ParameterizedDep{5},
		ParameterizedDep{6},
		ParameterizedDep{1},
		ParameterizedDep{1},
		ParameterizedDep{3},
		ParameterizedDep{6},
		ParameterizedDep{2},
	)
	mg.SerialDeps(
		ParameterizedDep{1},
		ParameterizedDep{2},
		ParameterizedDep{5},
	)
}

var Default = Main
