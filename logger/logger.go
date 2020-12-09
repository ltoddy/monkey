package logger

import (
	"fmt"
	"log"

	"github.com/ltoddy/monkey/colored"
)

type Logger struct {
	verbose bool
}

func New(verbose bool) *Logger {
	return &Logger{verbose: verbose}
}

func (logger *Logger) Println(format string, v ...interface{}) {
	if logger.verbose {
		log.Println(colored.Cyan(fmt.Sprintf(format, v...)))
	}
}

func (logger *Logger) Fatalln(format string, v ...interface{}) {
	log.Fatalln(colored.Red(fmt.Sprintf(format, v...)))
}
