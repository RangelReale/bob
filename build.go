package bob

import (
	"bytes"
	"fmt"
)

// MustBuild builds a query and panics on error
// useful for initializing queries that need to be reused
func MustBuild(q QueryWriter) (string, []any) {
	return MustBuildN(q, 1)
}

func MustBuildN(q QueryWriter, start int) (string, []any) {
	sql, args, err := BuildN(q, start)
	if err != nil {
		panic(err)
	}

	return sql, args
}

// Convinient function to build query from start
func Build(q QueryWriter) (string, []any, error) {
	return BuildN(q, 1)
}

// Convinient function to build query from a point
func BuildN(q QueryWriter, start int) (string, []any, error) {
	query, args, err := buildN(q, start)
	for _, arg := range args {
		if _, ok := arg.(NamedArgument); ok {
			return "", nil, fmt.Errorf("cannot use bob.NamedArgument with Build")
		}
	}
	return query, args, err
}

func buildN(q QueryWriter, start int) (string, []any, error) {
	b := &bytes.Buffer{}
	args, err := q.WriteQuery(b, start)

	return b.String(), args, err
}

func BuildPrepared(q QueryWriter) (PreparedQuery, error) {
	return BuildPreparedN(q, 1)
}

func BuildPreparedN(q QueryWriter, start int) (PreparedQuery, error) {
	query, args, err := buildN(q, start)
	if err != nil {
		return nil, err
	}

	var narg []NamedArgument
	for _, arg := range args {
		if na, ok := arg.(NamedArgument); ok {
			narg = append(narg, na)
		} else {
			return nil, fmt.Errorf("all arguments for BuildPrepared must be bob.NamedArgument")
		}
	}

	return preparedQuery{
		query: query,
		args:  narg,
	}, nil
}
