//go:build docs
// +build docs

package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"log"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/fatih/structtag"
	"golang.org/x/tools/go/packages"
)

const (
	interfacePackage = "github.com/oguzhand95/keeping-docs-in-sync-with-code/internal/config"
	interfaceName    = "Section"

	keyRequired        = "required"
	optionExampleValue = "example"
)

var (
	indentChar = strings.Repeat(" ", 4)

	header    = color.New(color.FgHiWhite, color.Bold).SprintFunc()
	fileName  = color.New(color.FgHiCyan).SprintfFunc()
	fieldDoc  = color.New(color.FgHiWhite).SprintfFunc()
	fieldName = color.New(color.FgHiGreen).SprintFunc()
	required  = color.New(color.FgHiWhite, color.Bold).SprintFunc()
	example   = color.New(color.FgHiYellow).SprintFunc()

	errTagNotExists = fmt.Errorf("tag does not exist")
)

func main() {
	absRootDir, err := filepath.Abs(".")
	if err != nil {
		log.Fatalf("Failed to find absolute path: %v", err)
	}

	pkgs, err := loadPackages(absRootDir)
	if err != nil {
		log.Fatalf("failed to load packages: %v", err)
	}

	iface, err := findInterfaceDef(pkgs)
	if err != nil {
		log.Fatalf("failed to find %s.%s: %v", interfacePackage, interfaceName, err)
	}

	log.Println(header("Conf structs"))
	structs := findIfaceImplementors(iface, pkgs)
	for _, s := range structs {
		log.Printf(fileName("%s.%s", s.Pkg, s.Name))
		if err = printFields(s.Fields, 1); err != nil {
			log.Fatalf("failed to print fields: %v", err)
		}
	}
}

func mkIndent(indent int) string {
	return strings.Repeat(indentChar, indent)
}

func printFields(fields []FieldInfo, indent int) error {
	for _, field := range fields {
		doc := ""

		tag, err := parseTag(field.Tag)
		if err != nil {
			return fmt.Errorf("failed to parse tags: %w", err)
		}

		if field.Documentation != "" {
			doc = fieldDoc("# %s", field.Documentation)
		}
		if tag.Required {
			doc = fmt.Sprintf("%s %s", doc, required("[REQUIRED]"))
		}

		if field.Fields != nil {
			log.Printf("%s%s: %s\n", mkIndent(indent), fieldName(tag.Name), doc)
			if field.Array {
				log.Printf("- \n")
				if err := printFields(field.Fields, indent+1); err != nil {
					return err
				}
				continue
			}
			if err := printFields(field.Fields, indent+1); err != nil {
				return err
			}
			continue
		}
		log.Printf("%s%s: %s %s", mkIndent(indent), fieldName(tag.Name), example(tag.DefaultValue), doc)
	}

	return nil
}

func loadPackages(pkgDir string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
		Dir:  pkgDir,
	}

	return packages.Load(cfg, "./...")
}

func findInterfaceDef(pkgs []*packages.Package) (*types.Interface, error) {
	for _, pkg := range pkgs {
		if obj := pkg.Types.Scope().Lookup(interfaceName); obj != nil {
			return obj.Type().Underlying().(*types.Interface), nil
		}
	}

	return nil, fmt.Errorf("failed to find the interface %s", interfaceName)
}

func implementsIface(iface *types.Interface, obj types.Object) bool {
	if obj == nil || !obj.Exported() {
		return false
	}

	t := obj.Type()
	if types.Implements(t, iface) {
		return true
	}

	ptr := types.NewPointer(t)
	if ptr != nil && types.Implements(ptr, iface) {
		return true
	}

	return false
}

func findIfaceImplementors(iface *types.Interface, pkgs []*packages.Package) []*StructInfo {
	var impls []*StructInfo
	for _, pkg := range pkgs {
		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			if name == interfaceName {
				continue
			}
			obj := scope.Lookup(name)
			if implementsIface(iface, obj) {
				if si := inspect(pkg, obj); si != nil {
					impls = append(impls, si)
				}
			}
		}
	}

	return impls
}

func inspect(pkg *packages.Package, obj types.Object) *StructInfo {
	var typespec *ast.TypeSpec

	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(node ast.Node) bool {
			ts, ok := node.(*ast.TypeSpec)
			if ok && ts.Name.Name == obj.Name() {
				typespec = ts
			}
			return true
		})
	}

	si := &StructInfo{Pkg: pkg.ID, Name: obj.Name()}
	si.Fields = inspectStruct(typespec.Type)

	return si
}

func inspectStruct(node ast.Expr) []FieldInfo {
	var fields []FieldInfo

	switch t := node.(type) {
	case *ast.StructType:
		for _, f := range t.Fields.List {
			if len(f.Names) == 0 {
				i, ok := f.Type.(*ast.Ident)
				if ok {
					ts, ok := i.Obj.Decl.(*ast.TypeSpec)
					if ok {
						st, ok := ts.Type.(*ast.StructType)
						if ok {
							fields = inspectStruct(st)
							continue
						}
					}
				}
			}

			fi := FieldInfo{Name: f.Names[0].Name, Documentation: strings.TrimSpace(f.Doc.Text())}
			if f.Tag != nil {
				fi.Tag = f.Tag.Value
			}

			if _, ok := f.Type.(*ast.ArrayType); ok {
				fi.Array = true
			}

			fi.Fields = inspectStruct(f.Type)

			fields = append(fields, fi)
		}
	case *ast.StarExpr:
		return inspectStruct(t.X)
	case *ast.Ident:
		if t.Obj != nil && t.Obj.Kind == ast.Typ {
			if ts, ok := t.Obj.Decl.(*ast.TypeSpec); ok {
				return inspectStruct(ts.Type)
			}
		}
	case *ast.ArrayType:
		return inspectStruct(t.Elt)
	}

	return fields
}

func parseTag(tag string) (*TagInfo, error) {
	t, err := structtag.Parse(tag[1 : len(tag)-1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse tags with structtag: %w", err)
	}

	if t == nil {
		return nil, errTagNotExists
	}

	yamlTag, err := t.Get("yaml")
	if err != nil {
		return nil, errTagNotExists
	}

	ti := &TagInfo{
		Name: yamlTag.Value(),
	}

	if confTag, _ := t.Get("conf"); confTag != nil {
		switch confTag.Name {
		case keyRequired:
			ti.Required = true
		}

		for _, option := range confTag.Options {
			sp := strings.SplitN(option, "=", 2)

			switch sp[0] {
			case optionExampleValue:
				ti.DefaultValue = sp[1]
			}
		}
	}

	return ti, nil
}

type StructInfo struct {
	Pkg    string
	Name   string
	Fields []FieldInfo
}

type FieldInfo struct {
	Name          string
	Documentation string
	Tag           string
	Fields        []FieldInfo
	Array         bool
}

type TagInfo struct {
	DefaultValue string
	Name         string
	Ignore       bool
	Required     bool
}
