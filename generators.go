package main

import (
	"fmt"
	"math"
	"math/rand"
)

const Operations string = "+-/*"

var OperationFunctions = map[uint8]interface{}{
	'+': func(a, b int) float64 {
		return float64(a + b)
	},
	'-': func(a, b int) float64 {
		return float64(a - b)
	},
	'*': func(a, b int) float64 {
		return float64(a * b)
	},
	'/': func(a, b int) float64 {
		return float64(a) / float64(b)
	},
}

func GenExpression() string {
	operation := Operations[rand.Intn(len(Operations))]
	operationFunc := OperationFunctions[operation].(func(int, int) float64)
	a := rand.Intn(100)
	b := rand.Intn(100)
	return fmt.Sprintf("%v %c %v = %v", a, operation, b, operationFunc(a, b))
}

func GenInequality() string {
	a := rand.Intn(100)
	b := rand.Intn(100)
	var c string
	if a > b {
		c = ">"
	} else if a < b {
		c = "<"
	} else {
		c = "="
	}
	return fmt.Sprintf("%v %v %v", a, c, b)
}

func GenSqrt() string {
	x := rand.Intn(100)
	return fmt.Sprintf("sqrt(%v) = %v", x, math.Sqrt(float64(x)))
}
