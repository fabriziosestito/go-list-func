package gen

import (
	"bytes"
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

type Import struct {
	Name string
	Path string
}

func GenerateStubs(patterns []string) error {
	pkgs, err := loadPackages(patterns)
	if err != nil {
		return err
	}

	_, err = getModuleName(".")
	if err != nil {
		return err
	}

	for _, pkg := range pkgs {
		err := os.MkdirAll(pkg.PkgPath, 0755)
		if err != nil {
			return err
		}

		buf := bytes.NewBuffer(nil)

		_, err = buf.WriteString("package " + pkg.Name + "\n\n")
		if err != nil {
			return err
		}
		// Get all the imports from the package and add it to the file
		// A the end we will programmatically use "goimports" on the generated file to fix the imports
		importedPackagesSet := make(map[string]struct{})
		for _, astFile := range pkg.Syntax {
			for _, o := range astFile.Imports {
				if isThirdParty(o.Path.Value) && !isLocalImport(o.Path.Value, pkgs) {
					continue
				}

				if o.Name != nil {
					if _, ok := importedPackagesSet[o.Name.Name]; ok {
						continue
					}

					_, err := buf.WriteString("import " + o.Name.Name + " " + o.Path.Value + "\n\n")
					if err != nil {
						return err
					}
					importedPackagesSet[o.Name.Name] = struct{}{}
				} else {
					name := o.Path.Value[strings.LastIndex(o.Path.Value, "/")+1:]
					name = strings.ReplaceAll(name, "\"", "")
					println(name)
					if _, ok := importedPackagesSet[name]; ok {
						continue
					}

					_, err := buf.WriteString("import " + o.Path.Value + "\n\n")
					if err != nil {
						return err
					}
					importedPackagesSet[name] = struct{}{}
				}
			}
		}

		importedPackages := []string{}
		for k := range importedPackagesSet {
			importedPackages = append(importedPackages, k)
		}

		for _, astFile := range pkg.Syntax {
			err = stubTypes(astFile, buf, importedPackages)
			if err != nil {
				return err
			}

			err = stubFunctions(astFile, buf, importedPackages)
			if err != nil {
				return err
			}

		}

		_, err = buf.WriteString("type Embedme interface{}\n\n")
		if err != nil {
			return (err)
		}

		// Programmatically use "goimports" on the generated file

		// The file is created before since the imports.Process() function
		// requires to know the file path.
		outFile, err := os.Create(pkg.PkgPath + "/" + pkg.Name + ".go")
		if err != nil {
			return err
		}

		res, _ := imports.Process(outFile.Name(), buf.Bytes(), nil)

		_, err = outFile.Write(res)
		if err != nil {
			return err
		}
	}

	return nil
}

// isThirdParty checks if the given import path is a third party package. (no standard library)
func isThirdParty(importPath string) bool {
	// Third party package import path usually contains "." (".com", ".org", ...)
	// This logic is taken from golang.org/x/tools/imports package.
	return strings.Contains(importPath, ".")
}

// isLocalImport checks if the given import path is local to the given packages.
func isLocalImport(importPath string, pkgs []*packages.Package) bool {
	for _, pkg := range pkgs {
		if "\""+pkg.PkgPath+"\"" == importPath {
			return true
		}
	}
	return false
}

// loadPackages loads packages from patterns.
func loadPackages(patterns []string) ([]*packages.Package, error) {
	var cfg packages.Config
	cfg.Mode |= packages.NeedName
	cfg.Mode |= packages.NeedSyntax

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

func stubTypes(astFile *ast.File, f *bytes.Buffer, importedPackages []string) error {
	for n, o := range astFile.Scope.Objects {
		// if o.Kind == ast.Typ {
		// check if type is exported(only need for non-local types)
		// if unicode.IsUpper([]rune(n)[0]) {
		node := o.Decl
		switch ts := node.(type) {
		case *ast.TypeSpec:
			switch t := ts.Type.(type) {
			case *ast.StructType:
				field := formatFieldsStruct(t.Fields, importedPackages)
				_, err := f.WriteString("type " + n + " struct " + "{" + field + "}\n\n")
				if err != nil {
					return err
				}
			case *ast.InterfaceType:
				i := "type " + n + " interface {\n"
				for _, method := range t.Methods.List {
					m, ok := method.Type.(*ast.FuncType)
					if !ok {
						// TODO: handle embedded interfaces
						continue
					}
					i += fmt.Sprintf("%s(%s) %s\n", method.Names[0].Name, formatFields(m.Params, importedPackages), formatFuncResults(m.Results, importedPackages))
				}
				i += "}\n\n"
				_, err := f.WriteString(i)
				if err != nil {
					return err
				}
			default:
				_, err := f.WriteString("type " + n + " " + formatTypeStruct(ts.Type, importedPackages) + "\n\n")
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func stubFunctions(astFile *ast.File, outFile *bytes.Buffer, importedPackages []string) error {
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

		foo := FormatFuncDecl(decl, importedPackages)
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
