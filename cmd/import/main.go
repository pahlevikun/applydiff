package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pahlevikun/applydiff/internal/cv/importer"
)

func main() {
	texPath := flag.String("tex", "cv/master_cv.tex", "input LaTeX resume")
	cvOut := flag.String("cv", "cv.json", "output cv.json path")
	factsOut := flag.String("facts", "facts.json", "output facts.json path")
	flag.Parse()

	abs, err := filepath.Abs(*texPath)
	if err != nil {
		log.Fatal(err)
	}
	doc, facts, err := importer.ImportTex(abs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "import warning: %v\n", err)
	}
	if err := importer.WriteJSON(*cvOut, *factsOut, doc, facts); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("wrote %s and %s\n", *cvOut, *factsOut)
}
