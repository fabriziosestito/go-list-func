# go-list-func

![Build Status](https://github.com/tony2001/go-list-func/workflows/Go/badge.svg)

go-list-func lists the functions in the package.

This package is a fork of [go-list-func](github.com/AkihiroSuda/go-list-func) reworked as a library with some additions.

## Examples
```
$ go-list-func -verbose -private ./...
func formatType(typ ast.Expr) string
func formatFields(fields *ast.FieldList) string
func formatFuncParams(fields *ast.FieldList) string
func formatFuncResults(fields *ast.FieldList) string
func FormatFuncDecl(pkgName string, decl *ast.FuncDecl) string
func FormatFuncDeclVerbose(pkgName string, decl *ast.FuncDecl) string
func LoadPackages(patterns []string, includeTests bool) ([]*packages.Package, error)
func WalkFuncs(pkgs []*packages.Package, applyFunc func(pkg *packages.Package, file *ast.File, decl *ast.FuncDecl) error) error
func IsExported(decl *ast.FuncDecl) bool
func isInterfaceDecl(decl *ast.FuncDecl) bool
func getRecvName(decl *ast.FuncDecl) string
func main()
func printFuncDecl(pkgName string, decl *ast.FuncDecl, verbose, private bool) error

```

```
$ go-list-func -verbose strings | sort
func Compare(a, b string) int
func Contains(s, substr string) bool
func ContainsAny(s, chars string) bool
func ContainsRune(s string, r rune) bool
func Count(s, sep string) int
func EqualFold(s, t string) bool
func Fields(s string) []string
func FieldsFunc(s string, f func(rune) bool) []string
func HasPrefix(s, prefix string) bool
func HasSuffix(s, suffix string) bool
func Index(s, sep string) int
func IndexAny(s, chars string) int
func IndexByte(s string, c byte) int
func IndexFunc(s string, f func(rune) bool) int
func IndexRune(s string, r rune) int
...
```

## CLI
```
Usage of ./go-list-func:
  -include-generated
        include generated files
  -include-tests
        include tests
  -print-package
        print package name for each function
  -private
        also print non-exported funcs
  -verbose
        verbose
```

## API
```go
// loads packages data for later usage
LoadPackages(patterns []string, includeTests bool) ([]*packages.Program, error)

// runs applyFunc for every function declaration in the package(s)
WalkFuncs(pkgs []*packages.Package, applyFunc func(pkg *packages.Package, file *ast.File, decl *ast.FuncDecl) error) error

// checks if the function/method is exported
IsExported(decl *ast.FuncDecl) bool

// format func declaration (short version)
FormatFuncDecl(pkgName string, decl *ast.FuncDecl) string

// format func declaration (verbose version)
FormatFuncDeclVerbose(pkgName string, decl *ast.FuncDecl) string
```
