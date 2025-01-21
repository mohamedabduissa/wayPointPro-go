package logging

import (
	"log"
	"os"
)

var Logger *log.Logger

func InitLogger() {
	Logger = log.New(os.Stdout, "API: ", log.LstdFlags|log.Lshortfile)
}
