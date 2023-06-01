package bob

import (
	"fmt"
)

type NamedArg struct {
	Name string
}

type NamedParams struct {
	params []any
}

func (n NamedParams) Params(values map[string]any) ([]any, error) {
	return n.getParams(values, true)
}

func (n NamedParams) ParamsNullable(values map[string]any) []any {
	params, _ := n.getParams(values, false)
	return params
}

func (n NamedParams) getParams(values map[string]any, errorIfNotSet bool) ([]any, error) {
	var ret []any
	for _, param := range n.params {
		var paramName string
		switch x := param.(type) {
		case NamedArg:
			paramName = x.Name
		case string:
			paramName = x
		default:
			if errorIfNotSet {
				return nil, fmt.Errorf("unsupported param type %T", param)
			}
			ret = append(ret, nil)
			continue
		}

		v, ok := values[paramName]
		if !ok {
			if errorIfNotSet {
				return nil, fmt.Errorf("parameter %s not set", paramName)
			}
			v = nil
		}

		ret = append(ret, v)
	}
	return ret, nil
}
