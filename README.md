# go-list-func

![Build Status](https://github.com/tony2001/go-list-func/workflows/Go/badge.svg)

go-list-func list up the list of functions in the package.

This package is a fork of [go-list-func](github.com/AkihiroSuda/go-list-func) reworked as a library with some additions.

## Example

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
    func Join(a []string, sep string) string
    func LastIndex(s, sep string) int
    func LastIndexAny(s, chars string) int
    func LastIndexByte(s string, c byte) int
    func LastIndexFunc(s string, f func(rune) bool) int
    func Map(mapping func(rune) rune, s string) string
    func NewReader(s string) *Reader
    func NewReplacer(oldnew string) *Replacer
    func Repeat(s string, count int) string
    func Replace(s, old, new string, n int) string
    func Split(s, sep string) []string
    func SplitAfter(s, sep string) []string
    func SplitAfterN(s, sep string, n int) []string
    func SplitN(s, sep string, n int) []string
    func Title(s string) string
    func ToLower(s string) string
    func ToLowerSpecial(_case unicode.SpecialCase, s string) string
    func ToTitle(s string) string
    func ToTitleSpecial(_case unicode.SpecialCase, s string) string
    func ToUpper(s string) string
    func ToUpperSpecial(_case unicode.SpecialCase, s string) string
    func Trim(s string, cutset string) string
    func TrimFunc(s string, f func(rune) bool) string
    func TrimLeft(s string, cutset string) string
    func TrimLeftFunc(s string, f func(rune) bool) string
    func TrimPrefix(s, prefix string) string
    func TrimRight(s string, cutset string) string
    func TrimRightFunc(s string, f func(rune) bool) string
    func TrimSpace(s string) string
    func TrimSuffix(s, suffix string) string
	...

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

Example:
```
$ ./go-list-func -verbose -private ./...
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
