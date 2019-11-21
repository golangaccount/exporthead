package exporthead

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
)

//var expoetOutReg = regexp.MustCompile(`^[A-Z]`)

func isExportGenDecl(decl *ast.GenDecl) bool {
	switch decl.Tok {
	case token.IMPORT:
		return true
	case token.CONST, token.VAR:
		if pass, v := isExportGenDeclValueSpecs(decl.Specs); pass {
			decl.Specs = v
			return true
		}
		return false
	case token.TYPE:
		if pass, v := isExportGenDeclTypeSpecs(decl.Specs); pass {
			decl.Specs = v
			return true
		}
		return false
	default:
		return false
	}
}

func isExportGenDeclValueSpecs(specs []ast.Spec) (bool, []ast.Spec) {
	var result []ast.Spec
	for _, item := range specs {
		if pass, r := isExportGenDeclValueSpec(item); pass {
			result = append(result, r)
		}
	}
	if len(result) == 0 {
		return false, nil
	}
	return true, result
}

func isExportGenDeclValueSpec(spec ast.Spec) (bool, ast.Spec) {
	if vs, ok := spec.(*ast.ValueSpec); ok {
		var names []*ast.Ident
		var values []ast.Expr

		for i, item := range vs.Names {
			if item.IsExported() {
				names = append(names, item)
				if vs.Values != nil {
					values = append(values, vs.Values[i])
				}
			}
		}
		if len(names) == 0 {
			return false, nil
		}
		vs.Names = names
		vs.Values = values
		return true, vs
	}
	return false, nil
}

func isExportGenDeclTypeSpecs(specs []ast.Spec) (bool, []ast.Spec) {
	var result []ast.Spec
	for _, item := range specs {
		if pass, v := isExportGenDeclTypeSpec(item); pass {
			result = append(result, v)
		}
	}
	if len(result) == 0 {
		return false, nil
	}
	return true, result
}

func isExportGenDeclTypeSpec(spec ast.Spec) (bool, ast.Spec) {
	//判断field 有export的就要export
	if ts, ok := spec.(*ast.TypeSpec); ok {
		fmt.Println(reflect.TypeOf(ts.Type))
		switch ts.Type.(type) {
		case *ast.BadExpr:
		case *ast.Ident:
			return ts.Name.IsExported(), ts
		case *ast.Ellipsis:
		case *ast.BasicLit:
		case *ast.FuncLit:
		case *ast.CompositeLit:
		case *ast.ParenExpr:
		case *ast.SelectorExpr:
		case *ast.IndexExpr:
		case *ast.SliceExpr:
		case *ast.TypeAssertExpr:
		case *ast.CallExpr:
		case *ast.StarExpr:
		case *ast.UnaryExpr:
		case *ast.BinaryExpr:
		case *ast.KeyValueExpr:

		case *ast.ArrayType:
			ts.Name.IsExported()
		case *ast.StructType:

		case *ast.FuncType:
			ts.Name.IsExported()
		case *ast.InterfaceType:
			ts.Name.IsExported()
		case *ast.MapType:
			ts.Name.IsExported()
		case *ast.ChanType:
			ts.Name.IsExported()
		default:
		}
	}
	return true, spec
}

// // An ArrayType node represents an array or slice type.
// ArrayType struct {
// 	Lbrack token.Pos // position of "["
// 	Len    Expr      // Ellipsis node for [...]T array types, nil for slice types
// 	Elt    Expr      // element type
// }

// // A StructType node represents a struct type.
// StructType struct {
// 	Struct     token.Pos  // position of "struct" keyword
// 	Fields     *FieldList // list of field declarations
// 	Incomplete bool       // true if (source) fields are missing in the Fields list
// }

// // Pointer types are represented via StarExpr nodes.

// // A FuncType node represents a function type.
// FuncType struct {
// 	Func    token.Pos  // position of "func" keyword (token.NoPos if there is no "func")
// 	Params  *FieldList // (incoming) parameters; non-nil
// 	Results *FieldList // (outgoing) results; or nil
// }

// // An InterfaceType node represents an interface type.
// InterfaceType struct {
// 	Interface  token.Pos  // position of "interface" keyword
// 	Methods    *FieldList // list of methods
// 	Incomplete bool       // true if (source) methods are missing in the Methods list
// }

// // A MapType node represents a map type.
// MapType struct {
// 	Map   token.Pos // position of "map" keyword
// 	Key   Expr
// 	Value Expr
// }

// // A ChanType node represents a channel type.
// ChanType struct {
// 	Begin token.Pos // position of "chan" keyword or "<-" (whichever comes first)
// 	Arrow token.Pos // position of "<-" (token.NoPos if there is no "<-")
// 	Dir   ChanDir   // channel direction
// 	Value Expr      // value type
// }
