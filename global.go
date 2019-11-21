package exporthead

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"

	"github.com/golangaccount/cmd.go.internal/load"
)

type scope struct {
	pkg    string
	name   string
	object *ast.Object
}

//存储当前包中的所有的scope变量(主要是用于进行快速的查找定位decl)
var packageScope = map[string]map[string]*scope{}

func addScope(pkg string, f *ast.File) {
	var pkgmap map[string]*scope
	if v, has := packageScope[pkg]; has {
		pkgmap = v
	} else {
		pkgmap = map[string]*scope{}
		packageScope[pkg] = pkgmap
	}
	for k, v := range f.Scope.Objects {
		if _, has := pkgmap[k]; has {
			panic(errors.New("重复定义"))
		}
		pkgmap[k] = &scope{
			pkg:    pkg,
			name:   k,
			object: v,
		}
	}
}
func getScope(pkg, name string) (bool, *scope) {
	if v, has := packageScope[pkg]; has {
		if s, has := v[name]; has {
			return true, s
		}
	}
	return false, nil
}

type decl struct {
	decl     ast.Decl   //struct
	relative []ast.Decl //依赖的decl 包括var struct func
}

func (d *decl) isExported() bool {
	return true
}

//pkg scope
var exportMark = map[string]map[string]*decl{}

func getDecl(pkgname, scope string) (bool, *decl) {
	if pkg, has := exportMark[pkgname]; has {
		if v, has := pkg[scope]; has {
			return has, v
		}
	}
	return false, nil
}

func setDecl(pkgname, scope string, value ast.Decl) *decl {
	if has, v := getDecl(pkgname, scope); has {
		return v
	}
	var pkg map[string]*decl
	if v, has := exportMark[pkgname]; has {
		pkg = v
	} else {
		pkg = map[string]*decl{}
		exportMark[pkgname] = pkg
	}
	if v, has := pkg[scope]; has {
		return v
	} else {
		pkg[scope] = &decl{
			decl: value,
		}
		return pkg[scope]
	}
}

//计算package
// func loadpackages(pkg []string) []*ast.File {
// 	load.Packages([]string{pkg})
// }
var fileset = token.NewFileSet()

func InitPkg(pkg *load.Package) {
	initPackage(pkg)
}

func initPackage(pkg *load.Package) ([]*ast.File, error) {
	result := make([]*ast.File, len(pkg.GoFiles))
	for i, item := range pkg.GoFiles {
		f, err := parser.ParseFile(fileset, filepath.Join(pkg.Dir, item), nil, parser.ParseComments)
		if err != nil {
			return result, err
		}
		result[i] = f
	}
	initPackageScope(pkg.ImportPath, result)
	initPackageDecl(pkg.ImportPath, result)
	return result, nil
}

func initPackageScope(pkg string, fs []*ast.File) {
	for _, item := range fs {
		addScope(pkg, item)
	}
}

//分析依赖
func initPackageDecl(pkg string, fs []*ast.File) {
	for _, item := range fs {
		for _, subitem := range item.Decls {
			switch subitem.(type) {
			case *ast.GenDecl:
				initPackageDeclGen(pkg, subitem.(*ast.GenDecl))
			case *ast.FuncDecl:
				initPakcageDeclFunc(pkg, subitem.(*ast.FuncDecl))
			}
		}
	}
}

func initPackageDeclGen(pkg string, gen *ast.GenDecl) {
	switch gen.Tok {
	case token.IMPORT: //不会产生依赖
	case token.CONST: //不会产生依赖
	case token.VAR:
		initPackageDeclGenVar(pkg, gen.Specs)
	case token.TYPE:
		initPackageDeclGenType(pkg, gen.Specs)
	}
}

func initPackageDeclGenVar(pkg string, specs []ast.Spec) {
	for _, item := range specs {
		switch item.(type) {
		case *ast.ValueSpec:
			initPackageDeclGenValueSpec(pkg, item.(*ast.ValueSpec))
		case *ast.TypeSpec:
			initPackageDeclGenTypeSpec(pkg, item.(*ast.TypeSpec))
		}
	}
}

func initPackageDeclGenType(pkg string, specs []ast.Spec) {

}

func initPackageDeclGenValueSpec(pkg string, spec *ast.ValueSpec) {
	v := reflect.ValueOf(spec.Type)
	if v.IsValid() && v.IsNil() {
		for _, item := range spec.Names {
			if item.IsExported() {
				initPackageDeclType(pkg, spec.Type)
				return
			}
		}
	} else if len(spec.Names) == len(spec.Values) {
		for i, item := range spec.Names {
			if item.IsExported() {
				initPackageDeclType(pkg, spec.Values[i])
			}
		}
	} else {
		//函数调用 进行函数解析，后再次调用该函数
	}

	// for i, item := range spec.Names {
	// 	if item.IsExported() {
	// 		//需要进行导出分析
	// 		fmt.Print(item.Name)
	// 		v := reflect.ValueOf(spec.Type)
	// 		if v.IsValid() && !v.IsNil() {
	// 			initPackageDeclType(pkg, spec.Type)
	// 		} else if len(spec.Names) == len(spec.Values) {
	// 			initPackageDeclType(pkg, spec.Values[i])
	// 		} else if len(spec.Values) == 1 { //func 多值返回
	// 			//需要计算func
	// 			if call, ok := spec.Values[0].(*ast.CallExpr); ok {
	// 				getCallExprResult(call)
	// 			}
	// 		}
	// 	}
	// }
}

func initPackageDeclGenValueSpecCall(pkg string, spec *ast.ValueSpec) {
	v := reflect.ValueOf(spec.Type)
	if v.IsValid() && v.IsNil() {
		for _, item := range spec.Names {
			if item.IsExported() {
				initPackageDeclType(pkg, spec.Type)
				return
			}
		}
	} else if len(spec.Names) == len(spec.Values) {
		for i, item := range spec.Names {
			if item.IsExported() {
				initPackageDeclType(pkg, spec.Values[i])
			}
		}
	} else {
		//函数调用 进行函数解析，后再次调用该函数
	}
}

func initPackageDeclGenTypeSpec(pkg string, spec *ast.TypeSpec) {
	fmt.Print(spec.Name)
	initPackageDeclType(pkg, spec.Type)
}

func initPakcageDeclFunc(pkg string, fun *ast.FuncDecl) {

}

//must 有export数据
func initPackageDeclType(pkg string, tp ast.Expr) {
	fmt.Println(reflect.TypeOf(tp))
	switch tp.(type) {
	case *ast.BadExpr:
	case *ast.Ident:
		//进行依赖添加
		//return ts.Name.IsExported(), ts
	case *ast.Ellipsis:
	case *ast.BasicLit:
	case *ast.FuncLit:
	case *ast.CompositeLit:
	case *ast.ParenExpr:
	case *ast.SelectorExpr: //skip package.Scope
	case *ast.IndexExpr:
	case *ast.SliceExpr:
	case *ast.TypeAssertExpr:
	case *ast.CallExpr: //func call
		initPackageDeclType(pkg, (tp.(*ast.CallExpr)).Fun)
	case *ast.StarExpr:
	case *ast.UnaryExpr:
	case *ast.BinaryExpr:
	case *ast.KeyValueExpr:

	case *ast.ArrayType:
		//ts.Name.IsExported()
	case *ast.StructType:

	case *ast.FuncType:
		//ts.Name.IsExported()
	case *ast.InterfaceType:
		//ts.Name.IsExported()
	case *ast.MapType:
		//ts.Name.IsExported()
	case *ast.ChanType:
		//ts.Name.IsExported()
	default:
	}
}

/******************************************************/
func getCallExprResult(call *ast.CallExpr) []ast.Expr {
	return nil
}

/******************************************************/
