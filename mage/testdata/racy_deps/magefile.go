// +build mage

package main

import (
	"fmt"
	"time"

	"github.com/magefile/mage/mg"
)

func Test1() error {
	fmt.Println("execute test1")
	return fmt.Errorf("error test1")
}

func Test2() error {
	mg.Deps(Test1)
	fmt.Println("execute test2")
	return nil
}

func Test3() error {
	time.Sleep(time.Second)
	mg.Deps(Test2)
	fmt.Println("execute test3")
	return nil
}
func Test4() error {
	mg.Deps(Test1, Test2, Test3)
	fmt.Println("execute test4")
	return nil
}
