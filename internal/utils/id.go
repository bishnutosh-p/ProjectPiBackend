package utils

import (
	"fmt"
	"time"
)

func GenerateUserID(num uint) string {
	return fmt.Sprintf("USER-%d-%d", num, time.Now().UnixNano())
}

func GenerateSongID(num uint) string {
	return fmt.Sprintf("SONG-%d-%d", num, time.Now().UnixNano())
}
