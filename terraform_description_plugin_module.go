package linters

import (
	"github.com/golangci/plugin-module-register/register"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("terraformdescription", New)
}

type MySettings struct {
	One   string    `json:"one"`
	Two   []Element `json:"two"`
	Three Element   `json:"three"`
}

type Element struct {
	Name string `json:"name"`
}

type PluginExample struct {
	settings MySettings
}

func New(settings any) (register.LinterPlugin, error) {
	// The configuration type will be map[string]any or []interface, it depends on your configuration.
	// You can use https://github.com/go-viper/mapstructure to convert map to struct.

	s, err := register.DecodeSettings[MySettings](settings)
	if err != nil {
		return nil, err
	}

	return &PluginExample{settings: s}, nil
}

func (f *PluginExample) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		{
			Name: "terraformdescription",
			Doc:  "checks that *schema.Schema has non-empty Description",
			Run:  f.run,
		},
	}, nil
}

func (f *PluginExample) GetLoadMode() string {
	return register.LoadModeSyntax
}

func (f *PluginExample) run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			cl, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}

			// Проверяем, что это schema.Schema
			if isSchemaType(cl.Type, "Schema") {
				checkDescriptionField(pass, cl)
				checkAttributesMap(pass, cl)
			}

			return true
		})
	}
	return nil, nil
}

func isSchemaType(expr ast.Expr, name string) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != name {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	return ok && ident.Name == "schema"
}

// Проверяет, есть ли Description в schema.Schema
func checkDescriptionField(pass *analysis.Pass, cl *ast.CompositeLit) {
	hasDescription := false
	for _, elt := range cl.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		if keyIdent, ok := kv.Key.(*ast.Ident); ok && (keyIdent.Name == "Description" || keyIdent.Name == "MarkdownDescription") {
			hasDescription = true
			break
		}
	}
	if !hasDescription {
		pass.Reportf(cl.Pos(), "schema.Schema should have Description or MarkdownDescription")
	}
}

// Проверяет, что в Attributes у каждого поля есть Description
func checkAttributesMap(pass *analysis.Pass, cl *ast.CompositeLit) {
	for _, elt := range cl.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		keyIdent, ok := kv.Key.(*ast.Ident)
		if !ok || keyIdent.Name != "Attributes" {
			continue
		}

		// Значение должно быть map[string]schema.Attribute
		attrMapLit, ok := kv.Value.(*ast.CompositeLit)
		if !ok {
			continue
		}

		for _, attr := range attrMapLit.Elts {
			attrKV, ok := attr.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			attrName, ok := attrKV.Key.(*ast.BasicLit) // "id", "name", ...
			if !ok || attrName.Kind != token.STRING {
				continue
			}

			// Значение — CompositeLit любого schema.*Attribute
			attrVal, ok := attrKV.Value.(*ast.CompositeLit)
			if !ok {
				continue
			}

			if !hasDescriptionField(attrVal) {
				pass.Reportf(attrKV.Pos(), "attribute %s is missing Description or MarkdownDescription", attrName.Value)
			}
		}
	}
}

func hasDescriptionField(cl *ast.CompositeLit) bool {
	for _, elt := range cl.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		k, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}
		if k.Name == "Description" || k.Name == "MarkdownDescription" {
			return true
		}
	}
	return false
}
