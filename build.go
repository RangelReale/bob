package bob

import (
	"bytes"
	"fmt"
)

// MustBuild builds a query and panics on error
// useful for initializing queries that need to be reused
func MustBuild(q Query) QueryBuilt {
	return MustBuildN(q, 1)
}

func MustBuildN(q Query, start int) QueryBuilt {
	qb, err := BuildN(q, start)
	if err != nil {
		panic(err)
	}

	return qb
}

// Convinient function to build query from start
func Build(q Query) (QueryBuilt, error) {
	return BuildN(q, 1)
}

// Convinient function to build query from a point
func BuildN(q Query, start int) (QueryBuilt, error) {
	b := &bytes.Buffer{}
	args, err := q.WriteQuery(b, start)
	if err != nil {
		return nil, err
	}

	var nargs []NamedArgument
	for _, arg := range args {
		if na, ok := arg.(NamedArgument); ok {
			nargs = append(nargs, na)
		} else if len(nargs) > 0 {
			return nil, fmt.Errorf("cannot mix named and non-named arguments")
		}
	}

	if len(nargs) > 0 {
		return &queryBuiltNamed{
			sql:   b.String(),
			args:  args,
			nargs: nargs,
		}, nil
	}

	return &queryBuiltDefault{
		sql:  b.String(),
		args: args,
	}, nil
}
