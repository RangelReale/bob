package bob

import (
	"fmt"
	"io"
)

func WithNamedArgs(q QueryWriter, args ...any) QueryWriter {
	return &queryWithNamedArgs{
		q:    q,
		args: args,
	}
}

type queryWithNamedArgs struct {
	q    QueryWriter
	args []any
}

func (q queryWithNamedArgs) WriteQuery(w io.Writer, start int) ([]any, error) {
	args, err := q.q.WriteQuery(w, start)
	if err != nil {
		return nil, err
	}

	var nargs []NamedArgument
	for _, arg := range args {
		if na, ok := arg.(NamedArgument); ok {
			nargs = append(nargs, na)
		} else {
			return nil, fmt.Errorf("cannot mix named and non-named arguments")
		}
	}

	return mergeNamedArguments(nargs, args...)
}

type buildResultDefault struct {
	sql  string
	args []any
}

func (q buildResultDefault) SQL() string {
	return q.sql
}

func (q buildResultDefault) Args() []any {
	return q.args
}

type buildResultNamed struct {
	sql   string
	args  []any
	nargs []NamedArgument
}

func (q buildResultNamed) SQL() string {
	return q.sql
}

func (q buildResultNamed) Args() []any {
	return q.args
}

func (q buildResultNamed) NamedArgs(args ...any) ([]any, error) {
	queryArgs, err := mergeNamedArguments(q.nargs, args...)
	if err != nil {
		return nil, err
	}

	return queryArgs, nil
}

func (q buildResultNamed) WithNamedArgs(args ...any) (BuildResult, error) {
	na, err := q.NamedArgs(args...)
	if err != nil {
		return nil, err
	}

	return &buildResultDefault{
		sql:  q.sql,
		args: na,
	}, nil
}
