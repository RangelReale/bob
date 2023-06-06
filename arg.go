package bob

import (
	"errors"
	"fmt"
)

type NamedArgument struct {
	Name string
}

func NamedArg(name string) NamedArgument {
	return NamedArgument{Name: name}
}

func namedArgumentMerge(nargs []NamedArgument, args any) ([]any, error) {
	var sourceArgs map[string]any

	switch a := args.(type) {
	case map[string]any:
		sourceArgs = a
	}

	// must try struct also

	if sourceArgs == nil {
		return nil, errors.New("unknown arguments type")
	}

	mergedArgs := make([]any, len(nargs))
	for idx, narg := range nargs {
		if carg, ok := sourceArgs[narg.Name]; ok {
			mergedArgs[idx] = carg
		} else {
			return nil, fmt.Errorf("named argument '%s' not found", narg.Name)
		}
	}

	return mergedArgs, nil
}

func NamesToNamedArguments(names ...string) []any {
	args := make([]any, len(names))
	for idx, name := range names {
		args[idx] = NamedArg(name)
	}
	return args
}
