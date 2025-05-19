package utils

import (
	"fmt"

	"github.com/fatih/color"
)

type terminal struct {
	ClearLines func(int)

	Lowkey func(string)
}

var Terminal terminal = terminal{
	ClearLines: clearLines,
	Lowkey:     lowkey,
}

const ansiMoveUp = "\033[A"
const ansiClearLine = "\033[2K"

func clearLines(n int) {
	for i := 0; i < n; i++ {
		fmt.Printf(ansiClearLine)
		fmt.Printf(ansiMoveUp)
	}
	fmt.Printf("\r")
}

func lowkey(s string) {
	color.New(color.FgHiBlack).Println(s)
}
