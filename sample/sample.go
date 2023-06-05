package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/scan"
)

func main() {
	main1()
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

	// SELECT
	// id, name
	// FROM users
	// WHERE ("id" IN ($1, $2, $3))
	//
	// [15 200 300]
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

	// INSERT INTO actor ("first_name", "last_name")
	// VALUES ($1, $2)
	//
	// [LAST_NAME 15]
}

func maindb() {
	db, err := sql.Open("pgx",
		fmt.Sprintf("postgres://postgres:password@%s:%s/%s?sslmode=disable", "localhost", "5478", "sakila"))
	if err != nil {
		panic(err)
	}

	type Data struct {
		FirstName string
		LastName  string
	}

	dataMapper := scan.StructMapper[Data]()

	for _, items := range [][2]int{{0, 4}, {2, 4}, {50, 12}} {
		fmt.Printf("%s OFFSET %d LIMIT %d %s", strings.Repeat("=", 10), items[0], items[1], strings.Repeat("=", 10))

		query := psql.Select(
			sm.Columns("first_name", "last_name"),
			sm.From("actor"),
			sm.OrderBy("first_name"),
			sm.OrderBy("last_name"),
			sm.Offset(psql.Arg(items[0])),
			sm.Limit(psql.Arg(items[1])),
		)
		sql, params, err := query.Build()
		if err != nil {
			panic(err)
		}
		fmt.Println(sql)
		fmt.Println(params)

		rows, err := db.Query(sql, params...)
		if err != nil {
			panic(err)
		}
		defer rows.Close()

		data, err := scan.AllFromRows(context.Background(), dataMapper, rows)
		if err != nil {
			panic(err)
		}

		fmt.Println(data)
	}

}
