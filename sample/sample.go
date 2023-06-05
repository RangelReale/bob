package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/scan"
)

func main() {
	main1()
}

func main1() {

}

func main2() {
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
