package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/cmd"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
)

func main() {

	cmd.RootCommand.Execute()

	rego.RegisterBuiltinDyn(
		&rego.Function{
			Name:    "redis",
			Decl:    types.NewVariadicFunction(types.Args(types.S), types.S, types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, terms []*ast.Term) (*ast.Term, error) {
			parsed := make([]string, 1)
			for _, t := range terms {
				var ter string
				if err := ast.As(t.Value, &ter); err != nil {
					return nil, err
				}
				ter = ""
				parsed = append(parsed, ter)
			}
			return nil, nil
		},
	)

	rego.RegisterBuiltin2(
		&rego.Function{
			Name:    "github.repo",
			Decl:    types.NewFunction(types.Args(types.S, types.S), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, a, b *ast.Term) (*ast.Term, error) {

			var org, repo string

			if err := ast.As(a.Value, &org); err != nil {
				return nil, err
			} else if ast.As(b.Value, &repo); err != nil {
				return nil, err
			}

			req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%v/%v", org, repo), nil)
			if err != nil {
				return nil, err
			}

			resp, err := http.DefaultClient.Do(req.WithContext(bctx.Context))
			if err != nil {
				return nil, err
			}

			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf(resp.Status)
			}

			v, err := ast.ValueFromReader(resp.Body)
			if err != nil {
				return nil, err
			}

			return ast.NewTerm(v), nil
		},
	)

	if err := cmd.RootCommand.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
