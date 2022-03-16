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
	"golang.org/x/tools/go/packages"
)

const (
	interfacePackage = "github.com/oguzhand95/keeping-docs-in-sync-with-code/internal/config"
	interfaceName    = "Section"
	indentChar       = "  "
)

var (
	Header    = color.New(color.FgHiWhite, color.Bold).SprintFunc()
	FileName  = color.New(color.FgHiCyan).SprintfFunc()
	FieldName = color.New(color.FgHiGreen).SprintFunc()
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

	log.Println(Header("Conf structs"))
	structs := findIfaceImplementors(iface, pkgs)
	for _, s := range structs {
		log.Printf(FileName("%s.%s", s.Pkg, s.Name))
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
		name := field.Name
		if field.Fields != nil {
			log.Printf("%s%s:\n", mkIndent(indent), name)
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
		log.Printf("%s%s ", mkIndent(indent), name)
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

	// TODO: Implement me
	// switch t := node.(type) {}

	return fields
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
