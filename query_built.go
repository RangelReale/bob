package bob

type queryBuiltDefault struct {
	sql  string
	args []any
}

func (q queryBuiltDefault) SQL() string {
	return q.sql
}

func (q queryBuiltDefault) Args() []any {
	return q.args
}

type queryBuiltNamed struct {
	sql   string
	args  []any
	nargs []NamedArgument
}

func (q queryBuiltNamed) SQL() string {
	return q.sql
}

func (q queryBuiltNamed) Args() []any {
	return q.args
}

func (q queryBuiltNamed) WithNamedArgs(args ...any) ([]any, error) {
	queryArgs, err := mergeNamedArguments(q.nargs, args...)
	if err != nil {
		return nil, err
	}

	return queryArgs, nil
}
