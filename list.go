package list

import (
	"fmt"
	"go/ast"
	"go/importer"
	"unicode"

	"golang.org/x/tools/go/packages"
)

func LoadPackages(patterns []string, includeTests bool) ([]*packages.Package, error) {
	var cfg packages.Config
	cfg.Mode |= packages.NeedName
	cfg.Mode |= packages.NeedSyntax
	cfg.Tests = includeTests

	pkgs, err := packages.Load(&cfg, patterns...)
	if err != nil {
		return nil, err
	}

	// packages.Load() returns a weird GRAPH-IN-ARRAY which means in can contain duplicates
	pkgMap := make(map[string]*packages.Package, len(pkgs))
	for _, pkg := range pkgs {
		pkgPath := pkg.PkgPath
		pkgMap[pkgPath] = pkg
	}

	pkgs = make([]*packages.Package, 0, len(pkgMap))
	for _, pkg := range pkgMap {
		pkgs = append(pkgs, pkg)
	}

	return pkgs, nil
}

type A []string

func WalkFuncs(pkgs []*packages.Package, applyFunc func(pkg *packages.Package, file *ast.File, decl *ast.FuncDecl) error) error {

	for _, pkg := range pkgs {
		imported, err := importer.Default().Import(pkg.Types.Path())
		if err != nil {
			panic(err)
		}
		for _, declName := range imported.Scope().Names() {
			fmt.Println(declName)
		}
		for _, file := range pkg.Syntax {
			for _, xdecl := range file.Decls {
				decl, ok := xdecl.(*ast.FuncDecl)
				if !ok {
					continue
				}

				if isInterfaceDecl(decl) {
					fmt.Println("interface")
					continue
				}

				if err := applyFunc(pkg, file, decl); err != nil {
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

// check if it's an interface method declaration
func isInterfaceDecl(decl *ast.FuncDecl) bool {
	if decl.Recv != nil {
		if len(decl.Recv.List) != 1 {
			panic(fmt.Errorf("strange receiver for %s: %#v", decl.Name.Name, decl.Recv))
		}

		field := decl.Recv.List[0]
		if len(field.Names) == 0 {
			return true
		}
	}
	return false
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
