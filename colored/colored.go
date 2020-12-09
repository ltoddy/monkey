package colored

import "fmt"

type Color uint8

const (
	ForegroundBlack Color = iota + 30
	ForegroundRed
	ForegroundGreen
	ForegroundYellow
	ForegroundBlue
	ForegroundMagenta
	ForegroundCyan
	ForegroundWhite
)

const (
	BackgroundBlack Color = iota + 40
	BackgroundRed
	BackgroundGreen
	BackgroundYellow
	BackgroundBlue
	BackgroundMagenta
	BackgroundCyan
	BackgroundWhite
)

const prefix = "\033["
const suffix = "\033[0m"

func colored(fg Color, content interface{}) string {
	return fmt.Sprintf("%s%vm%v%s", prefix, fg, content, suffix)
}

func coloredWithBackground(bg Color, fg Color, content interface{}) string {
	return fmt.Sprintf("%s%v;%vm%v%s", prefix, bg, fg, content, suffix)
}

func Black(content interface{}) string {
	return colored(ForegroundBlack, content)
}

func Red(content interface{}) string {
	return colored(ForegroundRed, content)
}

func Green(content interface{}) string {
	return colored(ForegroundGreen, content)
}

func Yellow(content interface{}) string {
	return colored(ForegroundYellow, content)
}

func Blue(content interface{}) string {
	return colored(ForegroundBlue, content)
}

func Magenta(content interface{}) string {
	return colored(ForegroundMagenta, content)
}

func Cyan(content interface{}) string {
	return colored(ForegroundCyan, content)
}

func White(content interface{}) string {
	return colored(ForegroundWhite, content)
}
