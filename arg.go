package bob

import (
	"errors"
	"fmt"
)

type ArgumentBinding struct {
	Name string
}

func ArgBinding(name string) ArgumentBinding {
	return ArgumentBinding{Name: name}
}

func replaceArgumentBindings(nargs []ArgumentBinding, argBinds BindSource) ([]any, error) {
	mergedArgs := make([]any, len(nargs))
	for idx, narg := range nargs {
		if carg, ok := argBinds.BindValue(narg.Name); ok {
			mergedArgs[idx] = carg
		} else {
			return nil, fmt.Errorf("argument binding '%s' not found", narg.Name)
		}
	}

	return mergedArgs, nil
}

func NamesToArgumentBindings(names ...string) []any {
	args := make([]any, len(names))
	for idx, name := range names {
		args[idx] = ArgBinding(name)
	}
	return args
}

func FailIfArgumentBindings(args []any) error {
	for _, arg := range args {
		if _, ok := arg.(ArgumentBinding); ok {
			return errors.New("some argument bindings were not processed")
		}
	}
	return nil
}

func FailIfMixedArgumentBindings(args []any) error {
	hasBinding := false
	hasNonBinding := false
	for _, arg := range args {
		if _, ok := arg.(ArgumentBinding); ok {
			hasBinding = true
		} else {
			hasNonBinding = true
		}
	}
	if hasBinding && hasNonBinding {
		return fmt.Errorf("cannot mix argument bindings with other arguments")
	}
	return nil
}
