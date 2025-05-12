package linters

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

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
	fmt.Println("Custom linter is running!")
	for _, file := range pass.Files {
		fmt.Printf("File %s analysis now \n", file.Name.Name)
		ast.Inspect(file, func(n ast.Node) bool {
			cl, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}

			// Проверяем, что это schema.Schema
			if isSchemaType(pass, cl, "Schema") {
				checkDescriptionField(pass, cl)
				checkAttributesMap(pass, cl)
			}

			return true
		})
	}
	return nil, nil
}

func isSchemaType(pass *analysis.Pass, cl *ast.CompositeLit, expected string) bool {
	t := pass.TypesInfo.Types[cl]
	if t.Type == nil {
		return false
	}
	named, ok := t.Type.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj.Pkg().Name() == "schema" && obj.Name() == expected
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

func checkAttributesMap(pass *analysis.Pass, cl *ast.CompositeLit) {
	fmt.Println("Start checkAttributesMap")
	for _, elt := range cl.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		keyIdent, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}

		if keyIdent.Name != "Attributes" {
			fmt.Printf("%s != Attributes", keyIdent.Name)
			continue
		}

		// Значение — map[string]schema.Attribute
		attrMap, ok := kv.Value.(*ast.CompositeLit)
		if !ok {
			continue
		}

		for _, attr := range attrMap.Elts {
			attrKV, ok := attr.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			attrName, ok := attrKV.Key.(*ast.BasicLit)
			if !ok || attrName.Kind != token.STRING {
				continue
			}

			attrVal, ok := attrKV.Value.(*ast.CompositeLit)
			if !ok {
				continue
			}

			checkAttributeLiteral(pass, attrName.Value, attrVal)
		}
	}
}

func checkAttributeLiteral(pass *analysis.Pass, name string, cl *ast.CompositeLit) {
	fmt.Println("Start checkAttributeLiteral")
	if !hasDescriptionField(cl) {
		pass.Reportf(cl.Pos(), "attribute %s is missing Description or MarkdownDescription", name)
	}

	for _, elt := range cl.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}

		switch key.Name {
		case "Attributes":
			inner, ok := kv.Value.(*ast.CompositeLit)
			if ok {
				for _, attr := range inner.Elts {
					innerKV, ok := attr.(*ast.KeyValueExpr)
					if !ok {
						continue
					}
					innerName, ok := innerKV.Key.(*ast.BasicLit)
					if !ok {
						continue
					}
					innerVal, ok := innerKV.Value.(*ast.CompositeLit)
					if !ok {
						continue
					}
					checkAttributeLiteral(pass, innerName.Value, innerVal)
				}
			}
		case "NestedObject":
			// for schema.ListNestedAttribute
			nestedObj, ok := kv.Value.(*ast.CompositeLit)
			if !ok {
				continue
			}
			for _, nestedElt := range nestedObj.Elts {
				nkv, ok := nestedElt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				if nkey, ok := nkv.Key.(*ast.Ident); ok && nkey.Name == "Attributes" {
					inner, ok := nkv.Value.(*ast.CompositeLit)
					if ok {
						for _, attr := range inner.Elts {
							innerKV, ok := attr.(*ast.KeyValueExpr)
							if !ok {
								continue
							}
							innerName, ok := innerKV.Key.(*ast.BasicLit)
							if !ok {
								continue
							}
							innerVal, ok := innerKV.Value.(*ast.CompositeLit)
							if !ok {
								continue
							}
							checkAttributeLiteral(pass, innerName.Value, innerVal)
						}
					}
				}
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
		fmt.Printf("hasDescriptionField(). Name: %s \n", k.Name)
		if k.Name == "Description" || k.Name == "MarkdownDescription" {
			return true
		}
	}
	return false
}
