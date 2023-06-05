package bob

import "io"

type PreparedQuery interface {
	Query(args any) QueryWriter
	Build(args any) (string, []any, error)
}

type preparedQuery struct {
	query string
	args  []NamedArgument
}

func (p preparedQuery) Query(args any) QueryWriter {
	queryArgs, err := NamedArgumentBuild(p.args, args)
	if err != nil {
		return &preparedQueryWriter{err: err}
	}
	return &preparedQueryWriter{
		query: p.query,
		args:  queryArgs,
	}
}

func (p preparedQuery) Build(args any) (string, []any, error) {
	queryArgs, err := NamedArgumentBuild(p.args, args)
	if err != nil {
		return "", nil, err
	}
	return p.query, queryArgs, nil
}

type preparedQueryWriter struct {
	query string
	args  []any
	err   error
}

func (p preparedQueryWriter) WriteQuery(w io.Writer, start int) ([]any, error) {
	_, err := w.Write([]byte(p.query))
	if err != nil {
		return nil, err
	}
	return p.args, nil
}
