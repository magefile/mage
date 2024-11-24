package deps

import (
	"fmt"

	"github.com/magefile/mage/mg"
)

// All code in this package belongs to @na4ma4 in GitHub https://github.com/na4ma4/magefile-test-import
// reproduced here for ease of testing regression on bug 508

type Docker mg.Namespace

func (Docker) Test() {
	fmt.Println("docker")
}

func Test() {
	fmt.Println("test")
}
