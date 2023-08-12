package bob

import (
	"errors"
	"fmt"
	"io"
)

func replaceArgumentBindingsWithSourceCheck(buildArgs []any, args ...any) ([]any, error) {
	var argBinds []BindSource
	for _, arg := range args {
		bs, ok := arg.(BindSource)
		if !ok {
			return nil, errors.New("argument must be BindSource")
		}
		argBinds = append(argBinds, bs)
	}
	return replaceArgumentBindingsWithCheck(buildArgs, &listBindSource{argBinds})
}

func replaceArgumentBindingsWithCheck(buildArgs []any, argBinds BindSource) ([]any, error) {
	var nargs []ArgumentBinding
	hasNonBinding := false
	for _, buildArg := range buildArgs {
		if na, ok := buildArg.(ArgumentBinding); ok {
			nargs = append(nargs, na)
		} else {
			hasNonBinding = true
		}
	}
	if len(nargs) == 0 {
		return buildArgs, nil
	}
	if hasNonBinding {
		return nil, fmt.Errorf("cannot mix argument bindings with other arguments")
	}
	return replaceArgumentBindings(nargs, argBinds)
}

type BindSource interface {
	BindValue(name string) (any, bool)
}

type mapBindSource struct {
	m map[string]any
}

func NewMapBindSource(m map[string]any) BindSource {
	return &mapBindSource{m}
}

type listBindSource struct {
	list []BindSource
}

func (l listBindSource) BindValue(name string) (any, bool) {
	for _, item := range l.list {
		if value, ok := item.BindValue(name); ok {
			return value, true
		}
	}
	return nil, false
}

func (m mapBindSource) BindValue(name string) (res any, ok bool) {
	res, ok = m.m[name]
	return
}

type BoundQuery interface {
	Query

	// MustBuild builds the query and panics on error
	// useful for initializing queries that need to be reused
	MustBuild() (string, []any)

	// MustBuildN builds the query and panics on error
	// start numbers the arguments from a different point
	MustBuildN(start int) (string, []any)

	// Convinient function to build query from start
	Build() (string, []any, error)

	// Convinient function to build query from a point
	BuildN(start int) (string, []any, error)
}

func BindQuery(q Query, argBinds BindSource) BoundQuery {
	return &boundQuery{
		q:        q,
		argBinds: argBinds,
	}
}

type boundQuery struct {
	q        Query
	argBinds BindSource
}

func (q boundQuery) WriteQuery(w io.Writer, start int) ([]any, error) {
	buildArgs, err := q.q.WriteQuery(w, start)
	if err != nil {
		return nil, err
	}
	return replaceArgumentBindingsWithCheck(buildArgs, q.argBinds)
}

func (q boundQuery) WriteSQL(w io.Writer, d Dialect, start int) ([]any, error) {
	buildArgs, err := q.q.WriteSQL(w, d, start)
	if err != nil {
		return nil, err
	}
	return replaceArgumentBindingsWithCheck(buildArgs, q.argBinds)
}

// MustBuild builds the query and panics on error
// useful for initializing queries that need to be reused
func (q boundQuery) MustBuild() (string, []any) {
	return MustBuild(q)
}

// MustBuildN builds the query and panics on error
// start numbers the arguments from a different point
func (q boundQuery) MustBuildN(start int) (string, []any) {
	return MustBuildN(q, start)
}

// Convinient function to build query from start
func (q boundQuery) Build() (string, []any, error) {
	return Build(q)
}

// Convinient function to build query from a point
func (q boundQuery) BuildN(start int) (string, []any, error) {
	return BuildN(q, start)
}
