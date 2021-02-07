package main

import (
	"algorithm/arithmetic"
	"fmt"
)

func main() {
	list := "[{\"key\":\"age\",\"value\":18}]"
	expression := "1-(2*age)+1"
	fmt.Println(arithmetic.Expression(expression, list))
}
