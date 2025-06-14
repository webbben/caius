package utils

import (
	"fmt"

	"github.com/fatih/color"
)

type terminal struct {
	ClearLines  func(int)
	ClearScreen func()
	Lowkey      func(string)
}

var Terminal terminal = terminal{
	ClearLines:  clearLines,
	ClearScreen: clearScreen,
	Lowkey:      lowkey,
}

const ansiMoveUp = "\033[A"
const ansiClearLine = "\033[2K"
const ansiClearScreen = "\033[2J\033[H"

func clearLines(n int) {
	for i := 0; i < n; i++ {
		fmt.Printf(ansiClearLine)
		fmt.Printf(ansiMoveUp)
	}
	fmt.Printf("\r")
}

func clearScreen() {
	fmt.Print(ansiClearScreen)
}

func lowkey(s string) {
	color.New(color.FgHiBlack).Println(s)
}
