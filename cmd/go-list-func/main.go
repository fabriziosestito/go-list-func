package main

import (
	"flag"
	"fmt"
	"go/ast"
	"os"

	"github.com/tony2001/go-list-func"
	"golang.org/x/tools/go/packages"
)

func main() {
	var (
		buildTags    string
		private      bool
		printPackage bool
		includeTests bool
		verbose      bool
	)

	flag.StringVar(&buildTags, "tags", "", "build tags")
	flag.BoolVar(&includeTests, "include-tests", false, "include tests")
	flag.BoolVar(&private, "private", false, "also print non-exported funcs")
	flag.BoolVar(&printPackage, "print-package", false, "print package name for each function")
	flag.BoolVar(&verbose, "verbose", false, "verbose")
	flag.Parse()

	pkgs, err := list.LoadPackages(flag.Args())
	if err != nil {
		fmt.Fprintf(os.Stderr, "LoadPackages(): %v\n", err)
		os.Exit(1)
	}

	applyFunc := func(pkg *packages.Package, file *ast.File, decl *ast.FuncDecl) error {
		pkgName := ""
		if printPackage {
			pkgName = pkg.Name
		}
		return printFuncDecl(pkgName, decl, verbose, private)
	}

	if err = list.WalkFuncs(pkgs, applyFunc); err != nil {
		fmt.Fprintf(os.Stderr, "WalkFuncs(): %v\n", err)
		os.Exit(1)
	}
}

func printFuncDecl(pkgName string, decl *ast.FuncDecl, verbose, private bool) error {
	if private || list.IsExported(decl) {
		if verbose {
			fmt.Println(list.FormatFuncDeclVerbose(pkgName, decl))
		} else {
			fmt.Println(list.FormatFuncDecl(pkgName, decl))
		}
	}

	return nil
}
