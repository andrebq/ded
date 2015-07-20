package main

import (
	"fmt"
	"github.com/google/gxui"
	"github.com/google/gxui/gxfont"
	"github.com/google/gxui/interval"
	"github.com/google/gxui/math"
	"github.com/google/gxui/mixins/base"
	"github.com/google/gxui/mixins/parts"
	"strings"
)

type (
	Direction gxui.KeyboardKey

	DedTextController struct {
		*gxui.TextBoxController
	}

	DedEditor struct {
		base.Control
		parts.Focusable
		parts.BackgroundBorderPainter

		controller *DedTextController

		theme gxui.Theme

		font gxui.Font

		scroll              math.Point
		tabWidthPlaceholder string

		oneLine bool
	}

	DedEditorOuter interface {
		base.ControlOuter
	}
)

const (
	DirLeft  = Direction(gxui.KeyLeft)
	DirRight = Direction(gxui.KeyRight)
	DirUp    = Direction(gxui.KeyUp)
	DirDown  = Direction(gxui.KeyDown)

	DirEol       = Direction(gxui.KeyEnd)
	DirBol       = Direction(gxui.KeyHome)
	DirBlockDown = Direction(gxui.KeyPageDown)
	DirBlockUp   = Direction(gxui.KeyPageUp)
)

func (de *DedEditor) Init(outer DedEditorOuter, theme gxui.Theme) {
	de.Control.Init(outer, theme)
	de.Focusable.Init(outer)
	de.controller = &DedTextController{gxui.CreateTextBoxController()}
	de.tabWidthPlaceholder = "        "
	de.theme = theme
}

func (de *DedEditor) GainedFocus() {
	de.Focusable.GainedFocus()
	de.Redraw()
}

func (de *DedEditor) LostFocus() {
	de.Focusable.LostFocus()
	de.Redraw()
}

func (de *DedEditor) Controller() *DedTextController {
	return de.controller
}

func (de *DedEditor) updateScroll(cv gxui.Canvas) {
	visibleRect := cv.Size().Rect()
	visibleRect = visibleRect.Offset(de.scroll)
	// the idea:
	// take the current position of the caret,
	// if it lives outside the visible area of the canvas (its size rect)
	// then, ajust as little as possible the viewport to display the cursor
	// force a new viewport

	// check if we can see the caret
	posTop, posBottom := de.caretPosition(false), de.caretPosition(true)
	if visibleRect.Contains(posBottom) && visibleRect.Contains(posTop) {
		return
	}

	// we measured using the top, but we need to see
	// both the top and the bottom, so consider the bottom
	// on the calculations
	switch {
	case posTop.Y < visibleRect.Min.Y:
		de.scroll.Y -= visibleRect.Min.Y - posTop.Y
	case posBottom.Y > visibleRect.Max.Y:
		de.scroll.Y += posBottom.Y - visibleRect.Max.Y
	}

	switch {
	case posTop.X < visibleRect.Min.X:
		de.scroll.X -= visibleRect.Min.X - posTop.X
	case posTop.X > visibleRect.Max.X:
		de.scroll.X += posTop.X - visibleRect.Max.X
	}

	return
}

func (de *DedEditor) caretPosition(showBottom bool) math.Point {
	caret := de.controller.LastCaret()
	line := de.controller.LineIndex(caret)

	lineHeight := de.LineHeight()
	column := de.MeasureRunes(de.controller.LineStart(line), caret).W

	if showBottom {
		line++
	}

	return math.Point{
		X: column,
		Y: lineHeight * line,
	}
}

func (de *DedEditor) projectRect(a math.Rect) math.Rect {
	ret := a.OffsetX(-de.scroll.X).OffsetY(-de.scroll.Y)
	return ret
}

func (de *DedEditor) projectRectY(a math.Rect) math.Rect {
	ret := a.OffsetY(-de.scroll.Y)
	return ret
}

func (de *DedEditor) Paint(cv gxui.Canvas) {
	de.updateScroll(cv)
	if de.HasFocus() {
		de.RenderCurrentLine(cv)
	}

	de.RenderText(cv)

	if de.HasFocus() {
		de.RenderCaret(cv)
	}
}

func (de *DedEditor) RenderText(cv gxui.Canvas) {
	for i := 0; i < de.controller.LineCount(); i++ {
		de.RenderSelection(cv, i)
		de.RenderLine(cv, i)
	}
}

func (de *DedEditor) RenderLine(cv gxui.Canvas, lineIdx int) {
	runes := de.controller.LineRunesTabReplace(de.controller.LineEnd(lineIdx), de.tabWidthPlaceholder)

	rect := cv.Size().Rect()
	lineHeight := de.LineHeight()
	rect = math.CreateRect(rect.Min.X, rect.Min.Y, rect.Max.X, lineHeight).OffsetY(lineHeight * lineIdx)
	rect = de.projectRect(rect)

	offsets := de.font.Layout(&gxui.TextBlock{
		Runes:     runes,
		AlignRect: rect,
		H:         gxui.AlignLeft,
		V:         gxui.AlignMiddle,
	})

	cv.DrawRunes(de.font, runes, offsets, gxui.White)
}

func (de *DedEditor) RenderSelection(cv gxui.Canvas, lineIdx int) {
	controller := de.controller
	ls, le := controller.LineStart(lineIdx), controller.LineEnd(lineIdx)

	offsetY := de.LineHeight() * lineIdx
	selections := controller.Selections()
	interval.Visit(&selections, gxui.CreateTextSelection(ls, le, false), func(s, e uint64, _ int) {
		x := de.measureRenderedLine(int(s)).W
		// measure the whole line
		m := de.measureRenderedLine(int(e))
		// and remove the non-selected part
		m.W -= x
		top := math.Point{X: x, Y: offsetY}
		bottom := top.Add(m.Point())
		de.PaintSelection(cv, top, bottom)
	})
}

func (de *DedEditor) PaintSelection(c gxui.Canvas, top, bottom math.Point) {
	r := math.Rect{Min: top, Max: bottom}
	r = de.projectRect(r)
	c.DrawRoundedRect(r, 1, 1, 1, 1, gxui.TransparentPen, gxui.Brush{Color: gxui.Gray40})
}

func (de *DedEditor) RenderCurrentLine(cv gxui.Canvas) {
	caret := de.controller.LastCaret()
	lineIdx := de.controller.LineIndex(caret)

	rect := cv.Size().Rect()
	lineHeight := de.LineHeight()
	rect = math.CreateRect(rect.Min.X, rect.Min.Y, rect.Max.X, lineHeight).OffsetY(lineHeight * lineIdx)
	blue := Color{gxui.Blue}.AddTransparency(0.8)

	// we need to fix only the y axis, since the
	// x is always at 0
	rect = de.projectRectY(rect)
	cv.DrawRect(rect, gxui.CreateBrush(blue.Color))
}

func (de *DedEditor) RenderCaret(cv gxui.Canvas) {
	caret := de.controller.LastCaret()
	lineIdx := de.controller.LineIndex(caret)

	rect := cv.Size().Rect()
	lineHeight := de.LineHeight()
	caretSize := de.CaretSize()

	rect = math.CreateRect(rect.Min.X, rect.Min.Y, rect.Min.X+caretSize.W, rect.Min.Y+caretSize.H)
	rect = rect.OffsetY(lineHeight * lineIdx)

	sz := de.measureRenderedLine(caret)
	rect = rect.OffsetX(sz.W - 1)

	rect = de.projectRect(rect)
	cv.DrawRect(rect, gxui.CreateBrush(gxui.White))
}

func (de *DedEditor) HandleEnter(ev gxui.KeyboardEvent) {
	controller := de.controller
	switch {
	case ev.Modifier.Control():
		// execute the selection
		controller.ExecuteSelection()
	default:
		controller.ReplaceWithNewlineKeepIndent()
	}
}

func (de *DedEditor) HandleMovement(dir Direction, isSelect, moveByWord bool) {
	controller := de.controller
	switch {
	case isSelect && moveByWord:
		switch dir {
		case DirLeft:
			controller.SelectLeftByWord()
		case DirRight:
			controller.SelectRightByWord()
		case DirUp:
			controller.SelectUp()
		case DirDown:
			controller.SelectDown()
		}
	case moveByWord:
		controller.ClearSelections()
		switch dir {
		case DirLeft:
			controller.MoveLeftByWord()
		case DirRight:
			controller.MoveRightByWord()
		case DirUp:
			controller.MoveUp()
		case DirDown:
			controller.MoveDown()
		}
	case isSelect:
		switch dir {
		case DirLeft:
			controller.SelectLeft()
		case DirRight:
			controller.SelectRight()
		case DirUp:
			controller.SelectUp()
		case DirDown:
			controller.SelectDown()
		case DirBol:
			controller.SelectHome()
		case DirEol:
			controller.SelectEnd()
		}
	default:
		controller.ClearSelections()
		switch dir {
		case DirLeft:
			controller.MoveLeft()
		case DirRight:
			controller.MoveRight()
		case DirUp:
			controller.MoveUp()
		case DirDown:
			controller.MoveDown()
		case DirBol:
			controller.MoveHome()
		case DirEol:
			controller.MoveEnd()
		case DirBlockUp, DirBlockDown:
			caret := controller.LastCaret()
			line := controller.LineIndex(caret)
			switch dir {
			case DirBlockDown:
				line += de.textBlockSize()
				if line > controller.LineCount() {
					line = controller.LineCount() - 1
				}
			case DirBlockUp:
				line -= de.textBlockSize()
				if line < 0 {
					line = 0
				}
			}
			caret = controller.LineStart(line)
			controller.SetCaret(caret)
		}
	}
}

func (de *DedEditor) KeyPress(ev gxui.KeyboardEvent) (consumed bool) {
	controller := de.controller
	consumed = true
	switch ev.Key {
	case gxui.KeyUp, gxui.KeyDown, gxui.KeyLeft, gxui.KeyRight,
		gxui.KeyHome, gxui.KeyEnd, gxui.KeyPageUp, gxui.KeyPageDown:
		de.HandleMovement(Direction(ev.Key), ev.Modifier.Shift(), ev.Modifier.Control())
	case gxui.KeyBackspace:
		controller.Backspace()
	case gxui.KeyDelete:
		controller.Delete()
	case gxui.KeyEnter:
		de.HandleEnter(ev)
	case gxui.KeyTab:
		controller.ReplaceAllRunes([]rune{'\t'})
		controller.ClearSelections()
	default:
		consumed = false
	}

	if consumed {
		// maybe we need to change something
		de.Redraw()
	}
	return
}

func (de *DedEditor) KeyStroke(ev gxui.KeyStrokeEvent) (consumed bool) {
	if !ev.Modifier.Control() && !ev.Modifier.Alt() {
		de.controller.ReplaceAllRunes([]rune{ev.Character})
		de.controller.ClearSelections()
		de.Redraw()
	}
	return true
}

func (de *DedEditor) Click(ev gxui.MouseEvent) (consumed bool) {
	caret := de.PointToTextCoord(ev.Point)
	de.controller.ClearSelections()
	de.controller.SetCaret(caret)
	de.Redraw()
	return
}

// PointToTextCoord returns a caret positon near the given pixelPostion
func (de *DedEditor) PointToTextCoord(pixelPoint math.Point) int {
	translatedPoint := pixelPoint.Add(de.scroll)
	line := translatedPoint.Y / de.LineHeight()

	if line >= de.controller.LineCount() {
		line = de.controller.LineCount() - 1
	}

	lineStart := de.controller.LineStart(line)
	lineEnd := de.controller.LineEnd(line)

	fullSize := de.measureRenderedLine(lineEnd)
	pixelPivot := float32(translatedPoint.X) / float32(fullSize.W)
	if pixelPivot > 1 {
		// the user clicked on the end of a smaller line
		return lineEnd
	}
	// the user clicked on pixelPivot units from the fullline
	// use this as a guess to reduce the number of checks needed

	pivot := int(float32(lineEnd-lineStart) * pixelPivot)

	for pivot > lineStart {
		sz := de.measureRenderedLine(pivot)
		if sz.W < translatedPoint.X {
			break
		}
		pivot--
	}

	return lineStart + pivot
}

func (de *DedEditor) LineHeight() int {
	sz := de.font.GlyphMaxSize()
	return sz.H
}

func (de *DedEditor) CaretSize() math.Size {
	sz := de.font.GlyphMaxSize()
	sz.W = 2
	return sz
}

func (de *DedEditor) MeasureRunes(start, end int) math.Size {
	return de.font.Measure(&gxui.TextBlock{
		Runes: de.controller.TextRunes()[start:end],
	})
}

func (de *DedEditor) Text() string {
	return de.controller.Text()
}

func (de *DedEditor) TextBytes() []byte {
	return []byte(de.Text())
}

func (de *DedEditor) SetTextBytes(txt []byte) {
	de.SetText(string(txt))
}

func (de *DedEditor) SetText(txt string) {
	if de.oneLine {
		replacer := strings.NewReplacer("\n", " ", "\n", " ")
		txt = replacer.Replace(txt)
	}
	de.controller.SelectAll()
	de.controller.SetText(txt)
	de.controller.ClearSelections()
}

func (de *DedEditor) SetFont(ttfData []byte, size int) {
	if ttfData == nil {
		ttfData = gxfont.Monospace
	}
	var err error
	de.font, err = de.theme.Driver().CreateFont(ttfData, size)
	if err == nil {
		de.font.LoadGlyphs(32, 126)
	} else {
		panic(fmt.Sprintf("Warning: Failed to load default monospace font - %v", err))
	}
}

func (de *DedEditor) measureRenderedLine(caret int) math.Size {
	return de.font.Measure(&gxui.TextBlock{
		Runes: de.controller.LineRunesTabReplace(caret, de.tabWidthPlaceholder),
	})
}

func (de *DedEditor) textBlockSize() int {
	return 10
}
