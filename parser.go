package mocky

/*
	Many of the functions in this file are (almost) verbatim copies from
	https://github.com/vektra/mockery.
*/

import (
	"fmt"
	"go/ast"
	"go/types"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

func Parse(dir string) ([]Interface, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatalf("failed to retrieve list of files: %s", err)
	}
	fpaths := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".go" {
			fpaths = append(fpaths, filepath.Join(dir, file.Name()))
		}
	}

	pkg, err := loadPackage(fpaths)
	if err != nil {
		return nil, err
	}

	interfaces := []Interface{}
	for i, file := range pkg.GoFiles {
		newInterfaces, err := parseFile(pkg, file, pkg.Syntax[i])
		if err != nil {
			return nil, err
		}
		interfaces = append(interfaces, newInterfaces...)
	}

	return interfaces, nil
}

func loadPackage(fpaths []string) (*packages.Package, error) {
	if len(fpaths) == 0 {
		return nil, fmt.Errorf("no file paths provided")
	}

	projectDir, err := getProjectRootDir(fpaths[0])
	if err != nil {
		return nil, err
	}

	mode := (packages.NeedTypes |
		packages.NeedFiles |
		packages.NeedSyntax |
		packages.NeedName |
		packages.NeedDeps)

	config := packages.Config{
		Mode: mode,
		Dir:  projectDir,
	}
	pkgs, err := packages.Load(&config, fpaths...)
	if err != nil {
		return nil, err
	}

	if len(pkgs) != 1 {
		return nil, fmt.Errorf("expected to find 1 package, found %d", len(pkgs))
	}

	return pkgs[0], nil
}

func getProjectRootDir(fpath string) (string, error) {
	const gomodFileName = "go.mod"

	dir := filepath.Dir(fpath)
	for len(dir) > 1 {
		if fileExists(path.Join(dir, gomodFileName)) {
			return dir, nil
		}
		dir = filepath.Dir(dir)
	}

	return "", fmt.Errorf("failed to find project root path: %w", os.ErrNotExist)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

type NodeVisitor struct {
	interfaceNames []string
}

func NewNodeVisitor() *NodeVisitor {
	return &NodeVisitor{
		interfaceNames: make([]string, 0),
	}
}

func (nv *NodeVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.TypeSpec:
		switch n.Type.(type) {
		case *ast.InterfaceType, *ast.FuncType:
			nv.interfaceNames = append(nv.interfaceNames, n.Name.Name)
		}
	}
	return nv
}

func parseFile(pkg *packages.Package, fpath string, syntax *ast.File) ([]Interface, error) {
	nv := NewNodeVisitor()
	ast.Walk(nv, syntax)

	interfaces := make([]Interface, 0, len(nv.interfaceNames))

	scope := pkg.Types.Scope()
	for _, ifaceName := range nv.interfaceNames {
		obj := scope.Lookup(ifaceName)
		if obj == nil {
			continue
		}

		typ, ok := obj.Type().(*types.Named)
		if !ok {
			continue
		}

		if typ.Obj().Pkg() == nil {
			continue
		}

		iface, ok := typ.Underlying().(*types.Interface)
		if !ok {
			continue
		}

		methods := make([]Method, iface.NumMethods())
		for i := 0; i < iface.NumMethods(); i++ {
			fn := iface.Method(i)
			signature := fn.Type().(*types.Signature)

			methods[i] = Method{
				Name:    fn.Name(),
				Args:    makeArguments(pkg.Types, signature.Params(), true),
				Returns: makeArguments(pkg.Types, signature.Results(), false),
			}
		}

		interfaces = append(interfaces, Interface{
			PackageName: pkg.Name,
			Name:        ifaceName,
			Methods:     methods,
		})
	}

	return interfaces, nil
}

func makeArguments(pkg *types.Package, tup *types.Tuple, mustHaveName bool) []Argument {
	args := make([]Argument, tup.Len())
	for n := 0; n < tup.Len(); n++ {
		param := tup.At(n)

		name := param.Name()
		if len(name) == 0 && mustHaveName {
			name = fmt.Sprintf("a%d", n)
		}
		args[n] = Argument{
			Name: name,
			Type: renderType(pkg, param.Type()),
		}
	}
	return args
}

func renderType(pkg *types.Package, typ types.Type) string {
	switch t := typ.(type) {
	case *types.Named:
		o := t.Obj()
		if o.Pkg() == nil || o.Pkg().Name() == "main" || o.Pkg() == pkg {
			return o.Name()
		}
		return o.Pkg().Name() + "." + o.Name()
	case *types.Basic:
		if t.Kind() == types.UnsafePointer {
			return "unsafe.Pointer"
		}
		return t.Name()
	case *types.Pointer:
		return "*" + renderType(pkg, t.Elem())
	case *types.Slice:
		return "[]" + renderType(pkg, t.Elem())
	case *types.Array:
		return fmt.Sprintf("[%d]%s", t.Len(), renderType(pkg, t.Elem()))
	case *types.Signature:
		switch t.Results().Len() {
		case 0:
			return fmt.Sprintf(
				"func(%s)",
				renderTypeTuple(pkg, t.Params()),
			)
		case 1:
			return fmt.Sprintf(
				"func(%s) %s",
				renderTypeTuple(pkg, t.Params()),
				renderType(pkg, t.Results().At(0).Type()),
			)
		default:
			return fmt.Sprintf(
				"func(%s)(%s)",
				renderTypeTuple(pkg, t.Params()),
				renderTypeTuple(pkg, t.Results()),
			)
		}
	case *types.Map:
		kt := renderType(pkg, t.Key())
		vt := renderType(pkg, t.Elem())

		return fmt.Sprintf("map[%s]%s", kt, vt)
	case *types.Chan:
		switch t.Dir() {
		case types.SendRecv:
			return "chan " + renderType(pkg, t.Elem())
		case types.RecvOnly:
			return "<-chan " + renderType(pkg, t.Elem())
		default:
			return "chan<- " + renderType(pkg, t.Elem())
		}
	case *types.Struct:
		var fields []string

		for i := 0; i < t.NumFields(); i++ {
			f := t.Field(i)

			if f.Anonymous() {
				fields = append(fields, renderType(pkg, f.Type()))
			} else {
				fields = append(fields, fmt.Sprintf("%s %s", f.Name(), renderType(pkg, f.Type())))
			}
		}

		return fmt.Sprintf("struct{%s}", strings.Join(fields, ";"))
	case *types.Interface:
		if t.NumMethods() != 0 {
			panic("Unable to mock inline interfaces with methods")
		}

		return "interface{}"
	default:
		panic(fmt.Sprintf("un-namable type: %#v (%T)", t, t))
	}
}

func renderTypeTuple(pkg *types.Package, tup *types.Tuple) string {
	var parts []string

	for i := 0; i < tup.Len(); i++ {
		v := tup.At(i)

		parts = append(parts, renderType(pkg, v.Type()))
	}

	return strings.Join(parts, " , ")
}
