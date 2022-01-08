package main

import (
	"fmt"
	"go/ast"
)

func formatType(typ ast.Expr) string {
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
		return fmt.Sprintf("func(%s)%s", formatFuncParams(t.Params), formatFuncResults(t.Results))
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", formatType(t.Key), formatType(t.Value))
	case *ast.ChanType:
		// FIXME
		panic(fmt.Errorf("unsupported chan type %#v", t))
	case *ast.BasicLit:
		return t.Value
	default:
		panic(fmt.Errorf("unsupported type %#v", t))
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
		s += formatType(field.Type)
		if i != len(fields.List)-1 {
			s += ", "
		}
	}

	return s
}

func formatFuncParams(fields *ast.FieldList) string {
	return formatFields(fields)
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

func formatFuncDecl(decl *ast.FuncDecl) string {
	s := "func "
	if decl.Recv != nil {
		if len(decl.Recv.List) != 1 {
			panic(fmt.Errorf("strange receiver for %s: %#v", decl.Name.Name, decl.Recv))
		}
		field := decl.Recv.List[0]
		if len(field.Names) == 0 {
			// function definition in interface (ignore)
			return ""
		} else if len(field.Names) != 1 {
			panic(fmt.Errorf("strange receiver field for %s: %#v", decl.Name.Name, field))
		}
		s += fmt.Sprintf("(%s %s) ", field.Names[0], formatType(field.Type))
	}

	s += fmt.Sprintf("%s(%s)", decl.Name.Name, formatFuncParams(decl.Type.Params))
	s += formatFuncResults(decl.Type.Results)

	return s
}
