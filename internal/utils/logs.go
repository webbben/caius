package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

func WriteLogs(content string) {
	f, err := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("failed to write logs:", err)
		return
	}
	defer f.Close()

	fmt.Fprintf(f, "---\n%s\n\n%s", time.Now(), content)
}
