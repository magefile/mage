package variadicimport

import "fmt"

// Collect prints imported variadic arguments.
func Collect(prefix string, args ...string) {
	fmt.Printf("shared:collect:%s:%q\n", prefix, args)
}
