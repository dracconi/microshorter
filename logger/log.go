package logger

import (
	"fmt"
	"time"
)

// Log s actions from other things into console
func Log(msg string) {
	fmt.Println("[" + time.Now().Format("15:04:05") + "] " + msg)
}
