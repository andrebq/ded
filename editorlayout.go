package main

import (
	"github.com/google/gxui"
	"github.com/google/gxui/math"
	"github.com/google/gxui/mixins/base"
)

type EditorLayoutOuter interface {
	base.ContainerOuter
}

// EditorLayout implements a panel where each child can have either a fixed
// size or can be expanded to fill the remaining space of the panel
type EditorLayout struct {
	base.Container
	theme gxui.Theme
	outer EditorLayoutOuter

	sizes map[gxui.Control]math.Size
}

func (l *EditorLayout) Init(outer EditorLayoutOuter, theme gxui.Theme) {
	l.Container.Init(outer, theme)
	l.outer = outer
	l.sizes = make(map[gxui.Control]math.Size)
}

func (l *EditorLayout) LayoutChildren() {
	s := l.outer.Size()
	children := l.outer.Children()
	if s == math.ZeroSize {
		return
	}

	// calculate the space that will be dynamic
	dynamic := s
	var numDynamics int
	for _, c := range children {
		if sz, ok := l.sizes[c.Control]; ok {
			dynamic.H -= sz.H
		} else {
			numDynamics++
		}
	}
	dynamicHeight := int(float32(dynamic.H) / float32(numDynamics))

	top := math.Point{}
	for _, c := range children {
		ctrl := c.Control
		var rect math.Rect

		if fixed, ok := l.sizes[ctrl]; ok {
			rect = math.CreateRect(top.X, top.Y, s.W, top.Y+fixed.H)
		} else {
			rect = math.CreateRect(top.X, top.Y, s.W, top.Y+dynamicHeight)
		}
		top.Y += rect.H()
		c.Layout(rect)
	}
}

func (l *EditorLayout) SetChildSize(c gxui.Control, sz math.Size) {
	if l.sizes[c] != sz {
		l.sizes[c] = sz
		l.LayoutChildren()
	}
}

func (l *EditorLayout) ChildSize(c gxui.Control) math.Size {
	return l.sizes[c]
}

func (l *EditorLayout) DesiredSize(min, max math.Size) math.Size {
	return max
}
