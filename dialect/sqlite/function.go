package sqlite

import (
	"io"

	"github.com/stephenafamo/bob/clause"
	"github.com/stephenafamo/bob/expr"
	"github.com/stephenafamo/bob/query"
)

type function struct {
	name string
	args []any

	// Used in value functions. Supported by Sqlite and Postgres
	filter []any

	// For chain methods
	expr.Chain[chain, chain]
}

// A function can be a target for a query
func (f *function) Apply(q *clause.FromItem) {
	q.Table = f
}

func (f *function) Filter(e ...any) *function {
	f.filter = append(f.filter, e...)

	return f
}

func (f *function) Over(window string) *functionOver {
	w := &functionOver{
		function: f,
		window:   clause.WindowDef{From: window},
	}
	w.Base = w
	return w
}

func (f function) WriteSQL(w io.Writer, d query.Dialect, start int) ([]any, error) {
	if f.name == "" {
		return nil, nil
	}

	w.Write([]byte(f.name))
	w.Write([]byte("("))
	args, err := query.ExpressSlice(w, d, start, f.args, "", ", ", "")
	if err != nil {
		return nil, err
	}
	w.Write([]byte(")"))

	filterArgs, err := query.ExpressSlice(w, d, start, f.filter, " FILTER (WHERE ", " AND ", ")")
	if err != nil {
		return nil, err
	}
	args = append(args, filterArgs...)

	return args, nil
}

type functionOver struct {
	function *function
	window   clause.WindowDef
	expr.Chain[chain, chain]
}

func (w *functionOver) PartitionBy(condition ...any) *functionOver {
	w.window.AddPartitionBy(condition...)
	return w
}

func (w *functionOver) OrderBy(order ...any) *functionOver {
	w.window.AddOrderBy(order...)
	return w
}

func (wr *functionOver) WriteSQL(w io.Writer, d query.Dialect, start int) ([]any, error) {
	fargs, err := query.Express(w, d, start, wr.function)
	if err != nil {
		return nil, err
	}

	winargs, err := query.ExpressIf(w, d, start+len(fargs), wr.window, true, "OVER (", ")")
	if err != nil {
		return nil, err
	}

	return append(fargs, winargs...), nil
}