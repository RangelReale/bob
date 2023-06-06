package bob

import (
	"bytes"
	"fmt"
)

// MustBuild builds a query and panics on error
// useful for initializing queries that need to be reused
func MustBuild(q QueryWriter) BuildResult {
	return MustBuildN(q, 1)
}

func MustBuildN(q QueryWriter, start int) BuildResult {
	qb, err := BuildN(q, start)
	if err != nil {
		panic(err)
	}

	return qb
}

// Convinient function to build query from start
func Build(q QueryWriter) (BuildResult, error) {
	return BuildN(q, 1)
}

// Convinient function to build query from a point
func BuildN(q QueryWriter, start int) (BuildResult, error) {
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
		return &buildResultNamed{
			sql:   b.String(),
			args:  args,
			nargs: nargs,
		}, nil
	}

	return &buildResultDefault{
		sql:  b.String(),
		args: args,
	}, nil
}
