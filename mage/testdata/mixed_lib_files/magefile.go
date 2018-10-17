//+build mage

package main

import "./subdir"

func Build() {
	subdir.Build()
}
