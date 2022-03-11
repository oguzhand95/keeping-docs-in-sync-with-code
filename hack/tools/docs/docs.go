//go:build docs
// +build docs

package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"log"
	"path/filepath"

	"github.com/fatih/color"
	"golang.org/x/tools/go/packages"
)

const (
	interfacePackage = "github.com/oguzhand95/keeping-docs-in-sync-with-code/internal/config"
	interfaceName    = "Section"
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

	log.Println(Header("FOUND STRUCTS"))
	structs := findIfaceImplementors(iface, pkgs)
	for _, structInfo := range structs {
		log.Printf(FileName("%s.%s", structInfo.Pkg, structInfo.Name))
		for _, field := range structInfo.Fields {
			log.Printf(FieldName(field.Name))
		}
	}
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

	// TODO: Implement
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
