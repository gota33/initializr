package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gota33/initializr/internal/assets"
)

var (
	name    = flag.String("name", "", "")
	pkgName = flag.String("package", "", "")
)

func main() {
	flag.Parse()

	baseName := strings.ToLower(*name)
	packageName := *pkgName
	workdir := filepath.Dir(os.Args[0])

	if packageName == "" {
		absName, _ := filepath.Abs(workdir)
		packageName = filepath.Base(absName)
	}

	tmpl, err := template.ParseFS(assets.FS, "*.tmpl")
	if err != nil {
		log.Fatalf("parse template: %s", err)
	}

	for _, subName := range strings.Split(baseName, ",") {
		if err = generate(subName, workdir, packageName, tmpl); err != nil {
			log.Fatal(err)
		}
	}
}

func generate(baseName, workdir, packageName string, tmpl *template.Template) (err error) {
	buf := &bytes.Buffer{}
	err = tmpl.ExecuteTemplate(buf, baseName+".tmpl", map[string]any{"PackageName": packageName})
	if err != nil {
		return fmt.Errorf("render template: %w", err)
	}

	outputName := filepath.Join(workdir, baseName+".go")
	err = os.WriteFile(outputName, buf.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	log.Printf("write to: %s", outputName)
	return
}
