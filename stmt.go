package bob

import (
	"context"
	"database/sql"

	"github.com/stephenafamo/scan"
)

type Preparer interface {
	Executor
	PrepareContext(ctx context.Context, query string) (Statement, error)
}

type Statement interface {
	ExecContext(ctx context.Context, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, args ...any) (scan.Rows, error)
}

// NewStmt wraps an [*sql.Stmt] and returns a type that implements [Queryer] but still
// retains the expected methods used by *sql.Stmt
// This is useful when an existing *sql.Stmt is used in other places in the codebase
func Prepare(ctx context.Context, exec Preparer, q QueryWriter) (Stmt, error) {
	sqlbuilt, err := Build(q)
	if err != nil {
		return Stmt{}, err
	}

	stmt, err := exec.PrepareContext(ctx, sqlbuilt.SQL())
	if err != nil {
		return Stmt{}, err
	}

	s := Stmt{
		executor: exec,
		stmt:     stmt,
		qb:       sqlbuilt,
	}

	if l, ok := q.(Loadable); ok {
		loaders := l.GetLoaders()
		s.loaders = make([]Loader, len(loaders))
		copy(s.loaders, loaders)
	}

	return s, nil
}

// Stmt is similar to *sql.Stmt but implements [Queryer]
type Stmt struct {
	stmt     Statement
	executor Executor
	qb       QueryBuilt
	loaders  []Loader
}

// Exec executes a query without returning any rows. The args are for any placeholder parameters in the query.
func (s Stmt) exec(ctx context.Context, args ...any) (sql.Result, error) {
	args, err := s.checkArgs(args...)
	if err != nil {
		return nil, err
	}

	result, err := s.stmt.ExecContext(ctx, args...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s Stmt) query(ctx context.Context, args ...any) (scan.Rows, error) {
	args, err := s.checkArgs(args...)
	if err != nil {
		return nil, err
	}

	rows, err := s.query(ctx, args...)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (s Stmt) checkArgs(args ...any) ([]any, error) {
	if qna, ok := s.qb.(QueryBuiltNamedArgs); ok {
		var err error
		args, err = qna.NamedArgs(args...)
		if err != nil {
			return nil, err
		}
	}
	return args, nil
}

func PrepareQuery[T any](ctx context.Context, exec Preparer, q QueryWriter, m scan.Mapper[T], opts ...ExecOption[T]) (QueryStmt[T, []T], error) {
	return PrepareQueryx[T, []T](ctx, exec, q, m, opts...)
}

func PrepareQueryx[T any, Ts ~[]T](ctx context.Context, exec Preparer, q QueryWriter, m scan.Mapper[T], opts ...ExecOption[T]) (QueryStmt[T, Ts], error) {
	var qs QueryStmt[T, Ts]

	s, err := Prepare(ctx, exec, q)
	if err != nil {
		return qs, err
	}

	settings := ExecSettings[T]{}
	for _, opt := range opts {
		opt(&settings)
	}

	if l, ok := q.(MapperModder); ok {
		if loaders := l.GetMapperMods(); len(loaders) > 0 {
			m = scan.Mod(m, loaders...)
		}
	}

	qs = QueryStmt[T, Ts]{
		stmt:     s,
		mapper:   m,
		settings: settings,
	}

	return qs, nil
}

type QueryStmt[T any, Ts ~[]T] struct {
	stmt     Stmt
	mapper   scan.Mapper[T]
	settings ExecSettings[T]
}

func (s QueryStmt[T, Ts]) Exec(ctx context.Context, args ...any) (sql.Result, error) {
	result, err := s.stmt.exec(ctx, args...)
	if err != nil {
		return nil, err
	}

	for _, loader := range s.stmt.loaders {
		if err := loader.Load(ctx, s.stmt.executor, nil); err != nil {
			return nil, err
		}
	}

	return result, err
}

func (s QueryStmt[T, Ts]) One(ctx context.Context, args ...any) (T, error) {
	var t T

	rows, err := s.stmt.query(ctx, args...)
	if err != nil {
		return t, err
	}

	t, err = scan.OneFromRows(ctx, s.mapper, rows)
	if err != nil {
		return t, err
	}

	for _, loader := range s.stmt.loaders {
		if err := loader.Load(ctx, s.stmt.executor, t); err != nil {
			return t, err
		}
	}

	if s.settings.AfterSelect != nil {
		if err := s.settings.AfterSelect(ctx, []T{t}); err != nil {
			return t, err
		}
	}

	return t, err
}

func (s QueryStmt[T, Ts]) All(ctx context.Context, args ...any) (Ts, error) {
	rows, err := s.stmt.query(ctx, args...)
	if err != nil {
		return nil, err
	}

	rawSlice, err := scan.AllFromRows(ctx, s.mapper, rows)
	if err != nil {
		return nil, err
	}

	typedSlice := Ts(rawSlice)

	for _, loader := range s.stmt.loaders {
		if err := loader.Load(ctx, s.stmt.executor, typedSlice); err != nil {
			return nil, err
		}
	}

	if s.settings.AfterSelect != nil {
		if err := s.settings.AfterSelect(ctx, typedSlice); err != nil {
			return nil, err
		}
	}

	return typedSlice, err
}

func (s QueryStmt[T, Ts]) Cursor(ctx context.Context, args ...any) (scan.ICursor[T], error) {
	rows, err := s.stmt.query(ctx, args...)
	if err != nil {
		return nil, err
	}

	m2 := scan.Mapper[T](func(ctx context.Context, c []string) (scan.BeforeFunc, func(any) (T, error)) {
		before, after := s.mapper(ctx, c)
		return before, func(link any) (T, error) {
			t, err := after(link)
			if err != nil {
				return t, err
			}

			for _, loader := range s.stmt.loaders {
				err = loader.Load(ctx, s.stmt.executor, t)
				if err != nil {
					return t, err
				}
			}

			if s.settings.AfterSelect != nil {
				if err := s.settings.AfterSelect(ctx, []T{t}); err != nil {
					return t, err
				}
			}
			return t, err
		}
	})

	return scan.CursorFromRows(ctx, m2, rows)
}
