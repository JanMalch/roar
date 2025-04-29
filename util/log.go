package util

import (
	"fmt"
	"io"

	"github.com/fatih/color"
)

var b = color.New(color.Bold)
var u = color.New(color.Underline)

// Logs an empty line. To be used for creating sections.
func LogEmptyLine(stdout io.Writer) {
	fmt.Fprintf(stdout, "\n")
}

func LogInfo(stdout io.Writer, format string, a ...any) {
	fmt.Fprintf(stdout, "%s %s\n", SymInfo, fmt.Sprintf(format, a...))
}

func LogSuccess(stdout io.Writer, format string, a ...any) {
	fmt.Fprintf(stdout, "%s %s\n", SymCheck, fmt.Sprintf(format, a...))
}

func LogWarning(stdout io.Writer, format string, a ...any) {
	fmt.Fprintf(stdout, "%s %s\n", SymWarn, fmt.Sprintf(format, a...))
}

func LogError(stderr io.Writer, format string, a ...any) {
	fmt.Fprintf(stderr, "%s %s\n", SymError, fmt.Sprintf(format, a...))
}

func Bold(s string) string {
	return b.Sprint(s)
}

func Underline(s string) string {
	return u.Sprint(s)
}

func Gray(s string) string {
	return color.HiBlackString(s)
}
