package exporthead

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"

	"github.com/golangaccount/cmd.go.internal/load"
	gos "github.com/golangaccount/go-libs/os"
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

//重写const，var
func rewriteValueSpec(spec *ast.ValueSpec) bool {
	specv := reflect.ValueOf(spec)
	if specv.IsValid() && !specv.IsNil() {
		result := make([]*ast.Ident, 0)
		for _, item := range spec.Names {
			if item.IsExported() {
				result = append(result, item)
			}
		}
		if len(result) == 0 {
			return false
		}
		spec.Names = result
		return true
	} else if len(spec.Names) == len(spec.Values) {
		result := make([]*ast.Ident, 0)
		values := make([]ast.Expr, 0)
		for i := 0; i < len(spec.Names); i++ {
			if spec.Names[i].IsExported() {
				result = append(result, spec.Names[i])
				values = append(values, spec.Values[i])
			}
		}
		if len(result) == 0 {
			return false
		}
		spec.Names = result
		spec.Values = values
		return true
	} else {
		mark := false
		for _, item := range spec.Names {
			if item.IsExported() {
				mark = true
			} else {
				item.Name = "_"
			}
		}
		return mark
	}
}

//重写decl
func rewriteGenDecl(decl *ast.GenDecl) bool {
	switch decl.Tok {
	case token.IMPORT:
		return true
	case token.CONST, token.VAR:
		result := make([]ast.Spec, 0)
		for i := 0; i < len(decl.Specs); i++ {
			if rewriteValueSpec(decl.Specs[i].(*ast.ValueSpec)) {
				result = append(result, decl.Specs[i])
			}
		}
		if len(result) == 0 {
			return false
		}
		decl.Specs = result
		return true
	case token.TYPE:
		return rewriteGenDeclType(decl)
	default:
		panic("")
	}
}

func rewriteGenDeclType(tp *ast.GenDecl) bool {
	//此处需要注意的是，type 默认是全部导出的，如果不全部导出的化需要分析对该type类型的依赖，包括 func、struct
	for _, item := range tp.Specs {
		spec := item.(*ast.TypeSpec)
		switch spec.Type.(type) {
		case *ast.StructType:
			rewriteFieldList((spec.Type.(*ast.StructType)).Fields)
		case *ast.InterfaceType:
			rewriteFieldList((spec.Type.(*ast.InterfaceType)).Methods)
		default:
		}
	}
	return true
}

//重写文件的decl
func rewriteFile(f *ast.File) {
	decls := make([]ast.Decl, 0)
	for i := 0; i < len(f.Decls); i++ {
		switch f.Decls[i].(type) {
		case *ast.GenDecl:
			if rewriteGenDecl(f.Decls[i].(*ast.GenDecl)) {
				decls = append(decls, f.Decls[i])
			}
		case *ast.FuncDecl:
			fd := f.Decls[i].(*ast.FuncDecl)
			if fd.Recv != nil {
				if fd.Name.IsExported() {
					decls = append(decls, fd)
				}
			} else {
				rewriteFunc(fd)
				decls = append(decls, fd)
			}
		}
	}
	f.Decls = decls
}

func ExportPackage(pkgs []string, dest string) {
	pkgsinfo := load.Packages(pkgs)
	files := make([]string, 0)
	dests := make([]string, 0)
	for _, item := range pkgsinfo {
		for _, file := range item.GoFiles {
			files = append(files, filepath.Join(item.Dir, file))
			dests = append(dests, filepath.Join(dest, item.ImportPath, file))
		}
	}
	ExportFile(files, dests)
}

func ExportFile(files []string, dest []string) {
	fset := token.NewFileSet()
	cfg := &Config{
		Mode:     UseSpaces | TabIndent,
		Tabwidth: 8,
	}
	for i, item := range files {
		if fs, err := parser.ParseFile(fset, item, nil, parser.ParseComments); err != nil {
			panic(err)
		} else {
			rewriteFile(fs)
			writeFileDisk(cfg, fset, fs, dest[i])
		}

	}
}

func writeFileDisk(cfg *Config, fset *token.FileSet, f *ast.File, dest string) {
	fs, err := gos.Create(dest)
	if err != nil {
		panic(err)
	}
	defer fs.Close()
	cfg.Fprint(fs, fset, f)
}
