package exporthead

import (
	"go/ast"
	"go/token"
)

//重写struct或interface,去掉无法导出部分
func rewriteFieldList(fl *ast.FieldList) {
	if fl.List == nil || len(fl.List) == 0 {
		return
	}
	result := make([]*ast.Field, 0)
	for _, item := range fl.List {
		idents := make([]*ast.Ident, 0)
		for _, item := range item.Names {
			if item.IsExported() {
				idents = append(idents, item)
			}
		}
		if len(idents) > 0 {
			item.Names = idents
			result = append(result, item)
		}
	}
	fl.List = result
}

//重写函数的body部分，去掉函数实现
func rewriteFunc(f *ast.FuncDecl) {
	if f.Type.Results == nil || len(f.Type.Results.List) == 0 {
		return
	}

	result := make([]ast.Expr, 0)
	for _, item := range f.Type.Results.List {
		if item.Names != nil {
			for i := 0; i < len(item.Names); i++ {
				result = append(result, &ast.Ident{
					Name: item.Names[i].Name, //使用default
				})
			}
		} else {
			result = append(result, funcReturnDefault(item))
		}
	}
	f.Body.List = []ast.Stmt{
		&ast.ReturnStmt{
			Results: result,
		},
	}
}

func funcReturnDefault(field *ast.Field) ast.Expr {
	switch field.Type.(type) {
	case *ast.BadExpr:
		panic("") //
	case *ast.Ident:
		return funcReturnDefaultIdent(field.Type.(*ast.Ident))
	case *ast.Ellipsis:
		panic("") //return not allow ...
	case *ast.BasicLit:
		panic("") //false,"",0 etc....
	case *ast.FuncLit:
		panic("")
	case *ast.CompositeLit:
		panic("") //{}
	case *ast.ParenExpr:
		panic("") // ()
	case *ast.SelectorExpr:
		return funcReturnDefaultSelector(field.Type.(*ast.SelectorExpr))
	case *ast.IndexExpr:
		panic("") //下标访问
	case *ast.SliceExpr:
		return funcReturnDefaultPtr()
	case *ast.TypeAssertExpr:
		panic("") //类型断言
	case *ast.CallExpr:
		panic("") //函数调用
	case *ast.StarExpr:
		return funcReturnDefaultPtr()
	case *ast.UnaryExpr:
		panic("") //一元运算符
	case *ast.BinaryExpr:
		panic("") //二元运算符
	case *ast.KeyValueExpr:
		panic("") //key value 键值对

	case *ast.ArrayType:
		return funcReturnDefaultArray(field.Type.(*ast.ArrayType))
	case *ast.StructType:
		panic("") //不可以出现类型定义
	case *ast.FuncType:
		return funcReturnDefaultPtr()
	case *ast.InterfaceType:
		return funcReturnDefaultPtr()
	case *ast.MapType:
		return funcReturnDefaultPtr()
	case *ast.ChanType:
		return funcReturnDefaultPtr()
	default:
	}
	return nil
}

func funcReturnDefaultIdent(ident *ast.Ident) ast.Expr {
	switch ident.Name {
	case "bool":
		return &ast.Ident{
			Name: "false", //使用default
		}
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64", "complex64", "complex128":
		return &ast.BasicLit{
			Kind:  5,
			Value: "0",
		}
	case "string":
		return &ast.BasicLit{
			Kind:  token.STRING,
			Value: "\"\"",
		}
	case "rune":
		return &ast.BasicLit{
			Kind:  token.CHAR,
			Value: "'0'",
		}
	default:
		return &ast.CompositeLit{
			Type: &ast.Ident{
				Name: ident.Name,
			},
		}
	}
}

func funcReturnDefaultPtr() ast.Expr {
	return &ast.Ident{
		Name: "nil", //使用default
	}
}

func funcReturnDefaultArray(array *ast.ArrayType) ast.Expr {
	return &ast.CompositeLit{
		Type: array,
	}
}

func funcReturnDefaultSelector(sel *ast.SelectorExpr) ast.Expr {
	return &ast.CompositeLit{
		Type: sel,
	}
}
