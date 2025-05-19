package utils

import "fmt"

type terminal struct {
	ClearLines func(int)
}

var Terminal terminal = terminal{
	ClearLines: clearLines,
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
