package linters

import (
	"go/ast"
	"go/token"

	"github.com/golangci/plugin-module-register/register"
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
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		composite, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		// Проверка типа: *schema.Schema
		selExpr, ok := composite.Type.(*ast.SelectorExpr)
		if !ok || selExpr.Sel.Name != "Schema" {
			return true
		}

		ident, ok := selExpr.X.(*ast.Ident)
		if !ok || ident.Name != "schema" {
			return true
		}

		for _, elt := range composite.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			keyIdent, ok := kv.Key.(*ast.Ident)
			if !ok || keyIdent.Name != "Description" {
				continue
			}

			val, ok := kv.Value.(*ast.BasicLit)
			if !ok || val.Kind != token.STRING {
				continue
			}

			if val.Value == `""` {
				pass.Reportf(val.Pos(), "Description in schema.Schema is empty")
			}
		}

		return true
	})

	return nil, nil
}
