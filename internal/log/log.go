package log

import (
	"io"
	"log"
)

const prefix = "[Initializr] "

var (
	std     = log.New(io.Discard, prefix, log.LstdFlags|log.Lmsgprefix)
	Printf  = std.Printf
	Println = std.Println
)

func SetOutput(w io.Writer) { std.SetOutput(w) }
