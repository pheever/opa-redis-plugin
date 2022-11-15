package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-redis/redis/v9"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/cmd"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
)

func main() {
	var client *redis.Client
	var addr string

	//get redis address from env vars - set default if not found
	addr, found := os.LookupEnv("OPA_REDIS_ADDR")
	if !found {
		addr = "redis://localhost:6379"
	}

	//parse address into redis opts - exit if error
	opts, err := redis.ParseURL(addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//create redis client and test connection
	client = redis.NewClient(opts)

	if _, e := client.Ping(context.Background()).Result(); e != nil {
		fmt.Println(e)
		os.Exit(1)
	}

	//implementation of redis command - get array of args and pass to redis client.Do - return result as string
	rediscmd := func(ctx rego.BuiltinContext, terms *ast.Term) (*ast.Term, error) {
		var args []interface{}
		ast.As(terms.Value, &args)

		p, err := client.Do(context.Background(), args[:]...).Result()

		return ast.StringTerm(fmt.Sprintf("%s", p)), err
	}

	//register opa 'redis' function
	rego.RegisterBuiltin1(
		&rego.Function{
			Name: "redis",
			Decl: types.NewFunction(types.Args(types.NewArray(nil, types.S)), types.S),
		},
		rediscmd,
	)

	//execute opa process
	if err := cmd.RootCommand.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
