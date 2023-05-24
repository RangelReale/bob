package psql_test

import (
	"testing"

	pg_query "github.com/pganalyze/pg_query_go/v4"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	testutils "github.com/stephenafamo/bob/test_utils"
)

func TestSelect(t *testing.T) {
	examples := testutils.Testcases{
		"simple select": {
			Doc:          "Simple Select with some conditions",
			ExpectedSQL:  "SELECT id, name FROM users WHERE (id IN ($1, $2, $3))",
			ExpectedArgs: []any{100, 200, 300},
			Query: psql.Select(
				sm.Columns("id", "name"),
				sm.From("users"),
				sm.Where(psql.Quote("id").In(psql.Arg(100, 200, 300))),
			),
		},
		"select distinct": {
			ExpectedSQL:  "SELECT DISTINCT id, name FROM users WHERE (id IN ($1, $2, $3))",
			ExpectedArgs: []any{100, 200, 300},
			Query: psql.Select(
				sm.Columns("id", "name"),
				sm.Distinct(),
				sm.From("users"),
				sm.Where(psql.Quote("id").In(psql.Arg(100, 200, 300))),
			),
		},
		"select distinct on": {
			ExpectedSQL:  "SELECT DISTINCT ON(id) id, name FROM users WHERE (id IN ($1, $2, $3))",
			ExpectedArgs: []any{100, 200, 300},
			Query: psql.Select(
				sm.Columns("id", "name"),
				sm.Distinct("id"),
				sm.From("users"),
				sm.Where(psql.Quote("id").In(psql.Arg(100, 200, 300))),
			),
		},
		"select from function": {
			Query: psql.Select(
				sm.From(psql.F("generate_series", 1, 3)).As("x", "p", "q", "s"),
			),
			ExpectedSQL:  `SELECT * FROM generate_series(1, 3) AS "x" ("p", "q", "s")`,
			ExpectedArgs: nil,
		},
		"with rows from": {
			Doc: "Select from group of functions. Automatically uses the `ROWS FROM` syntax",
			Query: psql.Select(
				sm.FromFunction(
					psql.F(
						"json_to_recordset",
						psql.Arg(`[{"a":40,"b":"foo"},{"a":"100","b":"bar"}]`),
					).Col("a", "INTEGER").Col("b", "TEXT"),
					psql.F("generate_series", 1, 3),
				).As("x", "p", "q", "s"),
				sm.OrderBy("p"),
			),
			ExpectedSQL: `SELECT *
				FROM ROWS FROM
					(
						json_to_recordset($1) AS (a INTEGER, b TEXT),
						generate_series(1, 3)
					) AS "x" ("p", "q", "s")
				ORDER BY p`,
			ExpectedArgs: []any{`[{"a":40,"b":"foo"},{"a":"100","b":"bar"}]`},
		},
		"with sub-select and window": {
			Doc: "Select from subquery with window function",
			ExpectedSQL: `SELECT status, avg(difference)
					FROM (
						SELECT
							status, 
							(LEAD(created_date, 1, NOW())
							OVER(PARTITION BY presale_id ORDER BY created_date)
							 - created_date) AS "difference"
						FROM presales_presalestatus
					) AS "differnce_by_status"
					WHERE status IN ('A', 'B', 'C')
					GROUP BY status`,
			Query: psql.Select(
				sm.Columns("status", psql.F("avg", "difference")),
				sm.From(psql.Select(
					sm.Columns(
						"status",
						psql.F("LEAD", "created_date", 1, psql.F("NOW")).
							Over("").
							PartitionBy("presale_id").
							OrderBy("created_date").
							Minus("created_date").
							As("difference")),
					sm.From("presales_presalestatus")),
				).As("differnce_by_status"),
				sm.Where(psql.Quote("status").In(psql.S("A"), psql.S("B"), psql.S("C"))),
				sm.GroupBy("status"),
			),
		},
		"select with grouped IN": {
			Query: psql.Select(
				sm.Columns("id", "name"),
				sm.From("users"),
				sm.Where(
					psql.Group(psql.Quote("id"), psql.Quote("employee_id")).
						In(psql.ArgGroup(100, 200), psql.ArgGroup(300, 400))),
			),
			ExpectedSQL:  "SELECT id, name FROM users WHERE (id, employee_id) IN (($1, $2), ($3, $4))",
			ExpectedArgs: []any{100, 200, 300, 400},
		},
		"function column with alias": {
			Doc:          "Function column with alias",
			ExpectedSQL:  `SELECT upper(name) AS "x" FROM users`,
			ExpectedArgs: nil,
			Query: psql.Select(
				sm.Columns(psql.F("upper", "name").As("x y")),
				sm.From("users"),
			),
		},
	}

	testutils.RunTests(t, examples, formatter)
}

func formatter(s string) (string, error) {
	aTree, err := pg_query.Parse(s)
	if err != nil {
		return "", err
	}

	return pg_query.Deparse(aTree)
}
