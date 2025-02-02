// pkg/utils/logger.go
package utils

import (
	"log"
)

func InitLogger() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
