package cmd

import (
	"fmt"
)

func Println(a ...any) {
	_, _ = fmt.Fprintln(app.Writer, a...)
}

func Printf(format string, a ...any) {
	_, _ = fmt.Fprintf(app.Writer, format, a...)
}

func Print(a ...any) {
	_, _ = fmt.Fprint(app.Writer, a...)
}

func ErrorLn(a ...any) {
	_, _ = fmt.Fprintln(app.ErrWriter, a...)
}

func Errorf(format string, a ...any) {
	_, _ = fmt.Fprintf(app.ErrWriter, format, a...)
}

func Error(a ...any) {
	_, _ = fmt.Fprint(app.ErrWriter, a...)
}
