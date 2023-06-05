package psql

import "github.com/stephenafamo/bob"

func NamedArg(name string) bob.NamedArgument {
	return bob.NamedArgument{
		Name: name,
	}
}
