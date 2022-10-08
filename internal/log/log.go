package log

import (
	"io"
	"log"
	"os"
)

const prefix = "[Initializr] "

var (
	std     = log.New(os.Stderr, prefix, log.LstdFlags|log.Lmsgprefix)
	Printf  = std.Printf
	Println = std.Println
	Fatalf  = std.Fatalf
	Fatal   = std.Fatal
)

func SetOutput(w io.Writer) { std.SetOutput(w) }
