package bob

type NamedArgument struct {
	Name string
}

func NamedArg(name string) NamedArgument {
	return NamedArgument{Name: name}
}
