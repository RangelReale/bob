package main

import (
	"fmt"

	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

func main() {
	main2()
}

func main1() {
	query := psql.Select(
		sm.Columns("id", "name"),
		sm.From("users"),
		sm.Where(psql.Quote("id").In(psql.Arg(psql.NamedArg("in1"), 200, 300))),
	)
	queryStr, params, err := query.BuildNamed()
	if err != nil {
		panic(err)
	}
	fmt.Println(queryStr)
	fmt.Println(params.ParamsNullable(map[string]any{
		"in1": 15,
	}))
}

func main2() {
	query := psql.Insert(
		im.Into("actor", "first_name", "last_name"),
		im.Values(psql.Arg("LAST_NAME", psql.NamedArg("in1"))),
	)
	queryStr, params, err := query.BuildNamed()
	if err != nil {
		panic(err)
	}
	fmt.Println(queryStr)
	fmt.Println(params.ParamsNullable(map[string]any{
		"in1": 15,
	}))
}
