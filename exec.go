package bob

import (
	"context"
	"database/sql"

	"github.com/stephenafamo/scan"
)

type (
	MapperModder interface {
		GetMapperMods() []scan.MapperMod
	}

	ExecSettings[T any] struct {
		AfterSelect func(ctx context.Context, retrieved []T) error
	}

	ExecOption[T any] func(*ExecSettings[T])
)

type Executor interface {
	scan.Queryer
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func Exec(ctx context.Context, exec Executor, q QueryWriter) (sql.Result, error) {
	sqlbuilt, err := Build(q)
	if err != nil {
		return nil, err
	}

	result, err := exec.ExecContext(ctx, sqlbuilt.SQL(), sqlbuilt.Args()...)
	if err != nil {
		return nil, err
	}

	if l, ok := q.(Loadable); ok {
		for _, loader := range l.GetLoaders() {
			if err := loader.Load(ctx, exec, nil); err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

func One[T any](ctx context.Context, exec Executor, q QueryWriter, m scan.Mapper[T], opts ...ExecOption[T]) (T, error) {
	settings := ExecSettings[T]{}
	for _, opt := range opts {
		opt(&settings)
	}

	var t T

	sqlbuilt, err := Build(q)
	if err != nil {
		return t, err
	}

	if l, ok := q.(MapperModder); ok {
		if loaders := l.GetMapperMods(); len(loaders) > 0 {
			m = scan.Mod(m, loaders...)
		}
	}

	t, err = scan.One(ctx, exec, m, sqlbuilt.SQL(), sqlbuilt.Args()...)
	if err != nil {
		return t, err
	}

	if l, ok := q.(Loadable); ok {
		for _, loader := range l.GetLoaders() {
			if err := loader.Load(ctx, exec, t); err != nil {
				return t, err
			}
		}
	}

	if settings.AfterSelect != nil {
		if err := settings.AfterSelect(ctx, []T{t}); err != nil {
			return t, err
		}
	}

	return t, err
}

func All[T any](ctx context.Context, exec Executor, q QueryWriter, m scan.Mapper[T], opts ...ExecOption[T]) ([]T, error) {
	return Allx[T, []T](ctx, exec, q, m, opts...)
}

// Allx takes 2 type parameters. The second is a special return type of the returned slice
// this is especially useful for when the the [Query] is [Loadable] and the loader depends on the
// return value implementing an interface
func Allx[T any, Ts ~[]T](ctx context.Context, exec Executor, q QueryWriter, m scan.Mapper[T], opts ...ExecOption[T]) (Ts, error) {
	settings := ExecSettings[T]{}
	for _, opt := range opts {
		opt(&settings)
	}

	sqlbuilt, err := Build(q)
	if err != nil {
		return nil, err
	}

	if l, ok := q.(MapperModder); ok {
		if loaders := l.GetMapperMods(); len(loaders) > 0 {
			m = scan.Mod(m, loaders...)
		}
	}

	rawSlice, err := scan.All(ctx, exec, m, sqlbuilt.SQL(), sqlbuilt.Args()...)
	if err != nil {
		return nil, err
	}

	typedSlice := Ts(rawSlice)

	if l, ok := q.(Loadable); ok {
		for _, loader := range l.GetLoaders() {
			if err := loader.Load(ctx, exec, typedSlice); err != nil {
				return typedSlice, err
			}
		}
	}

	if settings.AfterSelect != nil {
		if err := settings.AfterSelect(ctx, typedSlice); err != nil {
			return typedSlice, err
		}
	}

	return typedSlice, nil
}

// Cursor returns a cursor that works similar to *sql.Rows
func Cursor[T any](ctx context.Context, exec Executor, q QueryWriter, m scan.Mapper[T], opts ...ExecOption[T]) (scan.ICursor[T], error) {
	settings := ExecSettings[T]{}
	for _, opt := range opts {
		opt(&settings)
	}

	sqlbuilt, err := Build(q)
	if err != nil {
		return nil, err
	}

	if l, ok := q.(MapperModder); ok {
		if loaders := l.GetMapperMods(); len(loaders) > 0 {
			m = scan.Mod(m, loaders...)
		}
	}

	l, ok := q.(Loadable)
	if !ok {
		return scan.Cursor(ctx, exec, m, sqlbuilt.SQL(), sqlbuilt.Args()...)
	}

	m2 := scan.Mapper[T](func(ctx context.Context, c []string) (scan.BeforeFunc, func(any) (T, error)) {
		before, after := m(ctx, c)
		return before, func(link any) (T, error) {
			t, err := after(link)
			if err != nil {
				return t, err
			}

			for _, loader := range l.GetLoaders() {
				err = loader.Load(ctx, exec, t)
				if err != nil {
					return t, err
				}
			}

			if settings.AfterSelect != nil {
				if err := settings.AfterSelect(ctx, []T{t}); err != nil {
					return t, err
				}
			}
			return t, err
		}
	})

	return scan.Cursor(ctx, exec, m2, sqlbuilt.SQL(), sqlbuilt.Args()...)
}
