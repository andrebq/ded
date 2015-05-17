package main

import (
	"github.com/google/gxui"
)

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
