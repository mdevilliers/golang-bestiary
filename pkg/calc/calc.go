// package name: calc
package main

import (
	"C"
)

//export Sum
func Sum(x, y int) int {
	return x + y
}

func main() {
	// We need the main function to make possible
	// CGO compiler to compile the package as C shared library
}
