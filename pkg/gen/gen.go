package gen

import (
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	jijjo "encoding/json"

	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"
)

func GenerateStubs(patterns []string) error {
	_ = jijjo.Valid
	pkgs, err := loadPackages(patterns, false)
	if err != nil {
		return err
	}
	modName, err := getModuleName(".")
	if err != nil {
		return err
	}
	println("mod: ", modName)

	for _, pkg := range pkgs {
		err := os.MkdirAll(pkg.PkgPath, 0755)
		if err != nil {
			return err
		}

		outFile, err := os.Create(pkg.PkgPath + "/" + pkg.Name + ".go")
		if err != nil {
			return err
		}

		_, err = outFile.WriteString("package " + pkg.Name + "\n\n")
		if err != nil {
			return err
		}

		_, err = outFile.WriteString("type Embedme interface{}\n\n")
		if err != nil {
			return (err)
		}

		for _, astFile := range pkg.Syntax {
			for _, o := range astFile.Imports {
				for _, p := range pkgs {
					// println("p.PkgPath: ", p.PkgPath)

					// println("p.PkgPath: ", o.Path.Value)
					if "\""+p.PkgPath+"\"" == o.Path.Value || !isThirdParty(o.Path.Value) {
						if o.Name != nil {
							println(o.Name.Name)
						}
						println(o.Path.Value)
					}
				}
			}
			err = stubTypes(astFile, outFile)
			if err != nil {
				return err
			}

			err = stubFunctions(astFile, outFile)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func isThirdParty(importPath string) bool {
	// Third party package import path usually contains "." (".com", ".org", ...)
	// This logic is taken from golang.org/x/tools/imports package.
	return strings.Contains(importPath, ".")
}

func isLocalImport(importPath string, pkgs []*packages.Package) bool {
	for _, pkg := range pkgs {
		if pkg.PkgPath == importPath {
			return true
		}
	}
	return false
}

// loadPackages loads packages from patterns.
func loadPackages(patterns []string, includeTests bool) ([]*packages.Package, error) {
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

func stubTypes(astFile *ast.File, f *os.File) error {
	for n, o := range astFile.Scope.Objects {
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
					return err
				}
			default:
				_, err := f.WriteString("type " + n + " " + formatType(typeSpec.Type) + "\n\n")
				if err != nil {
					return err
				}
			}
		}
		// }
		// case *ast.InterfaceType:
	}

	return nil
}

func stubFunctions(astFile *ast.File, outFile *os.File) error {
	for _, xdecl := range astFile.Decls {
		decl, ok := xdecl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if isInterfaceDecl(decl) {
			continue
		}

		if !isExported(decl) {
			continue
		}

		foo := FormatFuncDecl("", decl)
		foo += " {\n panic(\"stub\")\n}\n\n"
		_, err := outFile.WriteString(foo)
		if err != nil {
			return err
		}
	}

	return nil
}

func isExported(decl *ast.FuncDecl) bool {
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

// getModuleName returns the Go module name from the go.mod file in the given directory
func getModuleName(dirPath string) (string, error) {
	goModFilePath := filepath.Join(dirPath, "go.mod")

	// Read the content of the go.mod file
	data, err := os.ReadFile(goModFilePath)
	if err != nil {
		return "", err
	}

	// Use modfile.ModulePath to extract the module name
	moduleName := modfile.ModulePath(data)
	if moduleName == "" {
		return "", fmt.Errorf("No module name found in go.mod")
	}

	return moduleName, nil
}
