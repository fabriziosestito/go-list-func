package list

import (
	"fmt"
	"go/ast"
	"go/build"
	"unicode"

	"golang.org/x/tools/go/loader"
)

func LoadProgram(tags, args []string, includeTests bool) (*loader.Program, error) {
	var conf loader.Config
	conf.Build = &build.Default
	conf.Build.BuildTags = append(conf.Build.BuildTags, tags...)

	_, err := conf.FromArgs(args, includeTests)
	if err != nil {
		return nil, err
	}

	prog, err := conf.Load()
	return prog, err
}

func WalkFuncsInProgram(prog *loader.Program, applyFunc func(decl *ast.FuncDecl) error) error {
	for _, pkgInfo := range prog.InitialPackages() {
		for _, file := range pkgInfo.Files {
			for _, xdecl := range file.Decls {
				decl, ok := xdecl.(*ast.FuncDecl)
				if !ok {
					continue
				}

				if err := applyFunc(decl); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func IsExported(decl *ast.FuncDecl) bool {
	isUpper0 := func(s string) bool {
		return unicode.IsUpper([]rune(s)[0])
	}

	if decl.Recv != nil {
		recvName := getRecvName(decl)
		return isUpper0(recvName) && isUpper0(decl.Name.Name)
	}

	return isUpper0(decl.Name.Name)
}

// get the name of the receiver struct
// Examples:
// func (b *Bar) testMethod() - Bar
// func (f foo) testMethod() - foo
func getRecvName(decl *ast.FuncDecl) string {
	// some new syntax?
	if len(decl.Recv.List) != 1 {
		panic(fmt.Errorf("multiple receivers for %s: %#v", decl.Name.Name, decl.Recv))
	}

	field := decl.Recv.List[0]
	// it has to be either "(s Some)" - Ident or "(s *Some)" - StarExpr
	switch t := field.Type.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		switch xType := t.X.(type) {
		case *ast.Ident:
			return xType.Name
		default:
			// not an identificator?
			panic(fmt.Errorf("unsupported receiver for %s: %#v", decl.Name.Name, decl.Recv))
		}

	default:
		// some new syntax?
		panic(fmt.Errorf("unsupported receiver for %s: %#v", decl.Name.Name, decl.Recv))
	}
}
