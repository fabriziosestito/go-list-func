package gen

import (
	"fmt"
	"go/ast"
	"go/token"
)

func formatType(typ interface{}) string {
	switch t := typ.(type) {
	case nil:
		return ""
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", formatType(t.X), t.Sel.Name)
	case *ast.StarExpr:
		return fmt.Sprintf("*%s", formatType(t.X))
	case *ast.ArrayType:
		return fmt.Sprintf("[%s]%s", formatType(t.Len), formatType(t.Elt))
	case *ast.Ellipsis:
		return formatType(t.Elt)
	case *ast.FuncType:
		return fmt.Sprintf("func(%s)%s", formatFields(t.Params), formatFuncResults(t.Results))
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", formatType(t.Key), formatType(t.Value))
	case *ast.ChanType:
		s := "chan"
		if t.Arrow != token.NoPos {
			if t.Begin == token.NoPos {
				s = "<- chan"
			} else {
				s = "chan <-"
			}
		}
		return fmt.Sprintf("%s %s", s, formatType(t.Value))
	case *ast.BasicLit:
		return t.Value
	case *ast.InterfaceType:
		return "interface {}"
	default:
		return ""
	}
}

func formatTypeStruct(typ interface{}) string {
	switch t := typ.(type) {
	case nil:
		return ""
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return "interface{}"
	case *ast.StarExpr:
		return fmt.Sprintf("*%s", formatTypeStruct(t.X))
	case *ast.ArrayType:
		return fmt.Sprintf("[%s]%s", formatTypeStruct(t.Len), formatTypeStruct(t.Elt))
	case *ast.Ellipsis:
		return formatTypeStruct(t.Elt)
	case *ast.FuncType:
		return fmt.Sprintf("func(%s)%s", formatFields(t.Params), formatFuncResults(t.Results))
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", formatTypeStruct(t.Key), formatTypeStruct(t.Value))
	case *ast.ChanType:
		s := "chan"
		if t.Arrow != token.NoPos {
			if t.Begin == token.NoPos {
				s = "<- chan"
			} else {
				s = "chan <-"
			}
		}
		return fmt.Sprintf("%s %s", s, formatTypeStruct(t.Value))
	case *ast.BasicLit:
		return t.Value
	case *ast.InterfaceType:
		return "interface {}"
	default:
		return ""
	}
}

func formatFields(fields *ast.FieldList) string {
	s := ""
	for i, field := range fields.List {
		for j, name := range field.Names {
			s += name.Name
			if j != len(field.Names)-1 {
				s += ","
			}
			s += " "
		}
		s += "interface{}"
		if i != len(fields.List)-1 {
			s += ", "
		}
	}

	return s
}

func formatFieldsStruct(fields *ast.FieldList) string {
	s := ""
	hasAlreadyEmbedded := false
	hasNamedFields := false
	for i, field := range fields.List {
		for j, name := range field.Names {

			s += name.Name
			if j != len(field.Names)-1 {
				s += ";"
			}
			s += " "
		}
		if len(field.Names) == 0 {
			if !hasAlreadyEmbedded {
				s += "Embedme"
				hasAlreadyEmbedded = true
			}
		} else {
			hasNamedFields = true
			s += formatTypeStruct(field.Type)
		}
		if i != len(fields.List)-1 && hasNamedFields {
			s += "; "
		}
	}

	println(s)
	return s
}

func formatFuncResults(fields *ast.FieldList) string {
	s := ""
	if fields != nil {
		s += " "
		if len(fields.List) > 1 {
			s += "("
		}
		s += formatFields(fields)
		if len(fields.List) > 1 {
			s += ")"
		}
	}

	return s
}

func FormatFuncDecl(pkgName string, decl *ast.FuncDecl) string {
	s := ""
	if pkgName != "" {
		s += fmt.Sprintf("%s: func ", pkgName)
	} else {
		s += "func "
	}

	if decl.Recv != nil {
		if len(decl.Recv.List) != 1 {
			panic(fmt.Errorf("strange receiver for %s: %#v", decl.Name.Name, decl.Recv))
		}

		field := decl.Recv.List[0]
		if len(field.Names) != 1 {
			panic(fmt.Errorf("strange receiver field for %s: %#v", decl.Name.Name, field))
		}
		s += fmt.Sprintf("(%s %s) ", field.Names[0], formatTypeStruct(field.Type))
	}

	s += fmt.Sprintf("%s(%s)", decl.Name.Name, formatFields(decl.Type.Params))
	s += formatFuncResults(decl.Type.Results)

	return s
}
