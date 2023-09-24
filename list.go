package list

import (
	"fmt"
	"go/ast"
	"os"
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
	// fmt.Printf("%v", pkgs)
	return pkgs, nil
}

type A []string

const B = 1

type C struct {
	banana A
	ciao   bool
}

func (c *C) testMethod() {
}

func WalkFuncs(pkgs []*packages.Package, applyFunc func(pkg *packages.Package, file *ast.File, decl *ast.FuncDecl) error) error {
	for _, pkg := range pkgs {
		err := os.MkdirAll(pkg.PkgPath, 0755)
		if err != nil {
			return err
		}

		f, err := os.Create(pkg.PkgPath + "/" + pkg.Name + ".go")
		if err != nil {
			return err
		}

		_, err = f.WriteString("package " + pkg.Name + "\n\nimport _ \"unsafe\"\n\n")
		if err != nil {
			return err
		}

		_, err = f.WriteString("type Embedme interface{}\n\n")
		if err != nil {
			return (err)
		}

		for _, file := range pkg.Syntax {

			generateTypes(f, file)
			for _, xdecl := range file.Decls {
				decl, ok := xdecl.(*ast.FuncDecl)
				if !ok {
					continue
				}

				if isInterfaceDecl(decl) {
					continue
				}

				// recv := ""
				// if decl.Recv != nil {
				// 	if len(decl.Recv.List) != 1 {
				// 		panic(fmt.Errorf("strange receiver for %s: %#v", decl.Name.Name, decl.Recv))
				// 	}

				// 	field := decl.Recv.List[0]
				// 	if len(field.Names) != 1 {
				// 		panic(fmt.Errorf("strange receiver field for %s: %#v", decl.Name.Name, field))
				// 	}
				// 	recv = fmt.Sprintf("(%s).", formatType(field.Type))
				// }

				// _, err := f.WriteString("//go:linkname stub_" + getRecvName(decl) + decl.Name.Name + " " + pkg.PkgPath + "." + recv + decl.Name.Name + "\n")
				// if err != nil {
				// 	return err
				// }

				// _, err = f.WriteString("func stub_" + getRecvName(decl) + decl.Name.Name + "() {\n    panic(\"stub\")\n}\n\n")
				// if err != nil {
				// 	return err
				// }

				foo := FormatFuncDeclVerbose("", decl)
				foo += " {\n panic(\"stub\")\n}\n\n"
				_, err := f.WriteString(foo)
				if err != nil {
					return err
				}
				// if err := applyFunc(pkg, file, decl); err != nil {
				// 	return err
				// }

			}
		}
	}
	return nil
}

func generateTypes(f *os.File, syntax *ast.File) {
	for n, o := range syntax.Scope.Objects {
		// if o.Kind == ast.Typ {
		// check if type is exported(only need for non-local types)
		// if unicode.IsUpper([]rune(n)[0]) {
		node := o.Decl
		switch node.(type) {
		case *ast.TypeSpec:
			typeSpec := node.(*ast.TypeSpec)
			switch typeSpec.Type.(type) {
			case *ast.StructType:
				structType := typeSpec.Type.(*ast.StructType)
				field := formatFieldsStruct(structType.Fields)
				_, err := f.WriteString("type " + n + " struct " + "{" + field + "}\n\n")
				if err != nil {
					panic(err)
				}
			default:
				_, err := f.WriteString("type " + n + " " + formatType(typeSpec.Type) + "\n\n")
				if err != nil {
					panic(err)
				}
			}
		}
		// case *ast.InterfaceType:
		// 	// n := typeSpec.Type.(*ast.ArrayType)
		// 	println(n + " is interface")

		// 	default:
		// 		println("type " + n + " " + formatType(typeSpec.Type))
		// 	}
		// case *ast.ValueSpec:
		// 	valueSpec := node.(*ast.ValueSpec)

		// 	println("var " + n + " " + formatType(valueSpec.Type))

		// 	for _, value := range valueSpec.Values {
		// 		println(formatType(value))
		// 	}
		// }

		// }
		// }
	}
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
	if decl.Recv == nil {
		return ""
	}
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
