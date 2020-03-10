package main

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
)

const pi float64 = 3.14

type Shape interface {
	getArea() float64
	getPerimeter() float64
}

type Circle struct {
	radius float64
}

func (c *Circle) getArea() float64      { return c.radius * c.radius * pi }
func (c *Circle) getPerimeter() float64 { return c.radius * pi * 2.0 }

type Rectangle struct {
	l float64
	r float64
}

func (rect Rectangle) getArea() float64       { return rect.l * rect.r }
func (rect *Rectangle) getPerimeter() float64 { return 2 * (rect.l + rect.r) }

func main() {
	const n int = 10
	arr := make([]Shape, n)
	for i := 0; i < n; i += 1 {
		val := common.RandInt() % 2
		if val == 0 {
			arr[i] = &Circle{
				radius: 10,
			}
		} else {
			arr[i] = &Rectangle{
				l: 10,
				r: 10,
			}
		}
	}
	fmt.Println(arr)
}
