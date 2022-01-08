package main

import (
	"flag"
	"fmt"
	"go/ast"
	"os"
	"strings"

	"github.com/tony2001/go-list-func"
)

func main() {
	var (
		buildTags    string
		private      bool
		includeTests bool
		verbose      bool
	)

	flag.StringVar(&buildTags, "tags", "", "build tags")
	flag.BoolVar(&includeTests, "include-tests", false, "include tests")
	flag.BoolVar(&private, "private", false, "also print non-exported funcs")
	flag.BoolVar(&verbose, "verbose", false, "verbose")
	flag.Parse()

	prog, err := list.LoadProgram(parseBuildTags(buildTags), flag.Args(), includeTests)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	applyFunc := func(decl *ast.FuncDecl) error {
		return printFuncDecl(decl, verbose, private)
	}

	if err = list.WalkFuncsInProgram(prog, applyFunc); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func parseBuildTags(tags string) []string {
	var result []string
	split := strings.Split(tags, ",")
	for _, s := range split {
		result = append(result, strings.TrimSpace(s))
	}

	return result
}

func printFuncDecl(decl *ast.FuncDecl, verbose, private bool) error {
	if private || list.IsExported(decl) {
		if verbose {
			fmt.Println(formatFuncDecl(decl))
		} else {
			fmt.Println(decl.Name.Name)
		}
	}

	return nil
}
