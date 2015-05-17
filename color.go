package main

import (
	"github.com/google/gxui"
)

type (
	Color struct {
		gxui.Color
	}
)

func (c Color) AddTransparency(val float32) Color {
	c.A -= val
	return c
}
