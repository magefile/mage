package mg_test

import (
	"fmt"

	"github.com/magefile/mage/mg"
)

func Example() {
	// Deps will run each dependency exactly once, and will run leaf-dependencies before those
	// functions that depend on them (if you put mg.Deps first in the function).

	// Normal (non-serial) Deps runs all dependencies in goroutines, so which one finishes first is
	// non-deterministic. Here we use SerialDeps here to ensure the example always produces the same
	// output.

	mg.SerialDeps(mg.F(Say, "hi"), Bark)
	// output:
	// hi
	// woof
}

func Say(something string) {
	fmt.Println(something)
}

func Bark() {
	mg.Deps(mg.F(Say, "woof"))
}
