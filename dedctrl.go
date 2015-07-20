package main

import (
	"amoraes.info/ded/commander"
	"github.com/google/gxui"
	"strings"
)

func (ctrl *DedTextController) ExecuteSelection() {
	selection := ctrl.LastSelection()
	selStart, selEnd := selection.Range()
	text := ctrl.TextRange(selStart, selEnd)

	if len(text) == 0 {
		// nothing selected, start from the current position
		// and walk backwards until the first whitespace
		// and walk forward to the next withespace or end of text
		runes := ctrl.TextRunes()

		lineStart := ctrl.LineStart(ctrl.LineIndex(selection.Start()))
		lineEnd := ctrl.LineEnd(ctrl.LineIndex(lineStart))

		// maybe we are in the end of the line,
		if selStart == lineEnd {
			selStart--
		}

		for selStart > lineStart {
			if runes[selStart] == ' ' {
				selStart++
				break
			}
			selStart--
		}

		for selEnd < lineEnd {
			if runes[selEnd] == ' ' {
				break
			}
			selEnd++
		}

		ctrl.ClearSelections()
		text = ctrl.TextRange(selStart, selEnd)
		selection = gxui.CreateTextSelection(selStart, selEnd, true)
		ctrl.SetSelection(selection)
	}

	c := &commander.C{}
	out, err := c.RunLine(text)
	if err != nil {
		ctrl.ReplaceAll(err.Error())
	} else {
		ctrl.ReplaceAll(strings.Trim(string(out), " \r\n\t"))
	}
}

func (ctrl *DedTextController) ReplaceTabWithSpace(tabwidth string) {
	var list gxui.TextSelectionList
	for i := 0; i < ctrl.LineCount(); i++ {
		// count how many tabs we have before any text
		runes := ctrl.LineRunes(i)
		for idx, runes := range runes {
			switch runes {
			case '\t':
				column := ctrl.LineStart(i) + idx
				list = append(list, gxui.CreateTextSelection(column, column+1, false))
			default:
				break
			}
		}
	}
	ctrl.SetSelections(list)

	ctrl.ReplaceAll(tabwidth)
}

// Use this only for rendering and measurements, since it will return the rune array
// with whitespace instead of tabs
func (ctrl *DedTextController) LineRunesTabReplace(caret int, tabwidth string) []rune {
	line := ctrl.LineIndex(caret)
	start, end := ctrl.LineStart(line), caret
	runes := ctrl.TextRunes()[start:end]
	var numtabs int
	for _, r := range runes {
		switch r {
		case '\t':
			numtabs++
		}
	}
	runeTab := gxui.StringToRuneArray(tabwidth)
	ret := make([]rune, 0, len(runes)+numtabs*len(runeTab))
	for _, r := range runes {
		switch r {
		case '\t':
			ret = append(ret, runeTab...)
		default:
			ret = append(ret, r)
		}
	}
	return ret
}
