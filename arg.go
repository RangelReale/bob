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

func NamedArgumentToArray(nargs []NamedArgument, args any) ([]any, error) {
	var argMap map[string]any

	switch xargs := args.(type) {
	// TODO: support struct
	case map[string]any:
		argMap = xargs
	}

	if argMap == nil {
		return nil, errors.New("unknown arguments type")
	}

	var retArgs []any
	for _, narg := range nargs {
		carg, ok := argMap[narg.Name]
		if !ok {
			return nil, fmt.Errorf("named argument '%s' not found", narg.Name)
		}
		retArgs = append(retArgs, carg)
	}

	return retArgs, nil
}
