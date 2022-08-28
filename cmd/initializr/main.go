package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gota33/initializr/internal/assets"
)

var (
	name    = flag.String("name", "", "")
	output  = flag.String("output", "", "")
	pkgName = flag.String("package", "", "")
)

func main() {
	flag.Parse()

	baseName := strings.ToLower(*name)
	outputName := *output
	packageName := *pkgName
	workdir := filepath.Dir(os.Args[0])

	if packageName == "" {
		absName, _ := filepath.Abs(workdir)
		packageName = filepath.Base(absName)
	}

	if outputName == "" {
		outputName = filepath.Join(workdir, baseName+".go")
	}

	tmpl, err := template.ParseFS(assets.FS, "*.tmpl")
	if err != nil {
		log.Fatalf("parse template: %s", err)
	}

	buf := &bytes.Buffer{}
	err = tmpl.ExecuteTemplate(buf, baseName+".tmpl", map[string]any{"PackageName": packageName})
	if err != nil {
		log.Fatalf("render template: %s", err)
	}

	err = os.WriteFile(outputName, buf.Bytes(), 0644)
	if err != nil {
		log.Fatalf("writing output: %s", err)
	}

	log.Printf("write to: %s", outputName)
}
