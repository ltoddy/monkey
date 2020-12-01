package logger

import "log"

type Logger struct {
	verbose bool
}

func New(verbose bool) *Logger {
	return &Logger{verbose: verbose}
}

func (logger *Logger) Print(v ...interface{}) {
	if logger.verbose {
		log.Print(v)
	}
}

func (logger *Logger) Printf(format string, v ...interface{}) {
	if logger.verbose {
		log.Printf(format, v)
	}
}

func (logger *Logger) Println(v ...interface{}) {
	if logger.verbose {
		log.Println(v)
	}
}

func (logger *Logger) Fatal(v ...interface{}) {
	log.Fatal(v)
}

func (logger *Logger) Fatalf(format string, v ...interface{}) {
	log.Fatalf(format, v)
}

func (logger *Logger) Fatalln(v ...interface{}) {
	log.Fatalln(v)
}
