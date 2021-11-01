package nested

import (
	"context"
	"fmt"
)

func Shutdown(context.Context) error {
	fmt.Println("Shutting down")
	return nil
}
