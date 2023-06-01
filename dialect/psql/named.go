package psql

import "github.com/stephenafamo/bob"

func NamedArg(name string) bob.NamedArg {
	return bob.NamedArg{
		Name: name,
	}
}
