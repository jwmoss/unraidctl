package api

import (
	"go/ast"
	"go/constant"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/vektah/gqlparser/v2"
	gqlast "github.com/vektah/gqlparser/v2/ast"
)

// Validate every string constant in queries.go, including composed operations.
func TestOperationsMatchReleasedSchemas(t *testing.T) {
	files := token.NewFileSet()
	file, err := parser.ParseFile(files, "queries.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	pkg, err := new(types.Config).Check("api", files, []*ast.File{file}, nil)
	if err != nil {
		t.Fatal(err)
	}
	fixtures, err := filepath.Glob("testdata/unraid-*.graphql")
	if err != nil || len(fixtures) == 0 {
		t.Fatalf("schema fixtures: %v", err)
	}
	for _, fixture := range fixtures {
		t.Run(filepath.Base(fixture), func(t *testing.T) {
			data, err := os.ReadFile(fixture)
			if err != nil {
				t.Fatal(err)
			}
			schema, err := gqlparser.LoadSchema(&gqlast.Source{Name: fixture, Input: string(data)})
			if err != nil {
				t.Fatal(err)
			}
			for _, model := range []struct {
				root  string
				field any
			}{
				{"docker", DockerResponse{}.Docker}, {"docker", DockerContainerResponse{}.Docker},
				{"docker", DockerLogsResponse{}.Docker}, {"vms", VMsResponse{}.VMs},
			} {
				field := schema.Query.Fields.ForName(model.root)
				if field == nil {
					t.Errorf("schema lacks the %s query root", model.root)
					continue
				}
				if !field.Type.NonNull && reflect.TypeOf(model.field).Kind() != reflect.Pointer {
					t.Errorf("%s is nullable and its response model must preserve null", model.root)
				}
			}
			for _, name := range pkg.Scope().Names() {
				value, ok := pkg.Scope().Lookup(name).(*types.Const)
				if !ok || value.Val().Kind() != constant.String {
					continue
				}
				t.Run(name, func(t *testing.T) {
					_, errs := gqlparser.LoadQueryWithRules(schema, constant.StringVal(value.Val()), nil)
					if len(errs) != 0 {
						t.Fatal(errs)
					}
				})
			}
		})
	}
}
