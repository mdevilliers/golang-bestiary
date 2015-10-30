// package: cashier
// filename: main.go
package main

import "github.com/mdevilliers/golang-bestiary/pkg/calc"
import "fmt"

func main() {
	fmt.Println("Cashier Application")
	fmt.Printf("Result: %d\n", calc.Sum(5, 10))
}
