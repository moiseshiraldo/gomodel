package gomodels

import (
	"strings"
	"testing"
)

// TestPostgresEngine tests the PostgresEngine methods
func TestPostgresEngine(t *testing.T) {
	model := &Model{
		name: "User",
		pk:   "id",
		fields: Fields{
			"id":      IntegerField{Auto: true},
			"email":   CharField{MaxLength: 100},
			"active":  BooleanField{DefaultFalse: true},
			"updated": DateTimeField{AutoNow: true},
		},
		meta: Options{
			Table:   "users_user",
			Indexes: Indexes{"test_index": []string{"email"}},
		},
	}
	mockedDB := &dbMocker{}
	engine := PostgresEngine{baseSQLEngine{
		db:          mockedDB,
		driver:      "postgres",
		escapeChar:  "\"",
		placeholder: "$",
	}}
	origScanRow := scanRow
	defer func() { scanRow = origScanRow }()
	scanRow = func(ex sqlExecutor, dest interface{}, query Query) error {
		db := ex.(*dbMocker)
		db.queries = append(db.queries, query)
		return nil
	}

	t.Run("CreateTable", func(t *testing.T) {
		mockedDB.Reset()
		if err := engine.CreateTable(model, false); err != nil {
			t.Fatal(err)
		}
		if len(mockedDB.queries) != 1 {
			t.Fatalf("expected one query, got %d", len(mockedDB.queries))
		}
		stmt := mockedDB.queries[0].Stmt
		if !strings.HasPrefix(stmt, `CREATE TABLE IF NOT EXISTS "users_user"`) {
			t.Errorf(
				"expected query start: %s",
				`CREATE TABLE IF NOT EXISTS "users_user"`,
			)
		}
		if !strings.Contains(stmt, `"email" VARCHAR(100) NOT NULL`) {
			t.Errorf(
				"expected query to contain: %s",
				`"email" VARCHAR(100) NOT NULL`,
			)
		}
		if !strings.Contains(stmt, `"id" SERIAL`) {
			t.Errorf("expected query to contain: %s", `"id" INTEGER`)
		}
	})

	t.Run("RenameTable", func(t *testing.T) {
		mockedDB.Reset()
		newModel := &Model{meta: Options{Table: "new_table"}}
		if err := engine.RenameTable(model, newModel); err != nil {
			t.Fatal(err)
		}
		if len(mockedDB.queries) != 1 {
			t.Fatalf("expected one query, got %d", len(mockedDB.queries))
		}
		expected := `ALTER TABLE "users_user" RENAME TO "new_table"`
		stmt := mockedDB.queries[0].Stmt
		if stmt != expected {
			t.Errorf("expected:\n\n%s\n\ngot:\n\n%s", expected, stmt)
		}
	})

	t.Run("DropTable", func(t *testing.T) {
		mockedDB.Reset()
		if err := engine.DropTable(model); err != nil {
			t.Fatal(err)
		}
		if len(mockedDB.queries) != 1 {
			t.Fatalf("expected one query, got %d", len(mockedDB.queries))
		}
		expected := `DROP TABLE "users_user"`
		stmt := mockedDB.queries[0].Stmt
		if stmt != expected {
			t.Errorf("expected:\n\n%s\n\ngot:\n\n%s", expected, stmt)
		}
	})

	t.Run("AddIndex", func(t *testing.T) {
		mockedDB.Reset()
		if err := engine.AddIndex(model, "test_index", "email"); err != nil {
			t.Fatal(err)
		}
		if len(mockedDB.queries) != 1 {
			t.Fatalf("expected one query, got %d", len(mockedDB.queries))
		}
		expected := `CREATE INDEX "test_index" ON "users_user" ("email")`
		stmt := mockedDB.queries[0].Stmt
		if stmt != expected {
			t.Errorf("expected:\n\n%s\n\ngot:\n\n%s", expected, stmt)
		}
	})

	t.Run("DropIndex", func(t *testing.T) {
		mockedDB.Reset()
		if err := engine.DropIndex(model, "test_index"); err != nil {
			t.Fatal(err)
		}
		if len(mockedDB.queries) != 1 {
			t.Fatalf("expected one query, got %d", len(mockedDB.queries))
		}
		expected := `DROP INDEX "test_index"`
		stmt := mockedDB.queries[0].Stmt
		if stmt != expected {
			t.Errorf("expected:\n\n%s\n\ngot:\n\n%s", expected, stmt)
		}
	})

	t.Run("AddColumns", func(t *testing.T) {
		mockedDB.Reset()
		fields := Fields{
			"is_superuser": BooleanField{},
			"created":      DateTimeField{AutoNowAdd: true},
		}
		if err := engine.AddColumns(model, fields); err != nil {
			t.Fatal(err)
		}
		if len(mockedDB.queries) != 1 {
			t.Fatalf("expected 1 queries, got %d", len(mockedDB.queries))
		}
		stmt := mockedDB.queries[0].Stmt
		if !strings.HasPrefix(stmt, `ALTER TABLE "users_user" ADD COLUMN`) {
			t.Errorf(
				"expected query start: %s",
				`ALTER TABLE "users_user" ADD COLUMN`,
			)
		}
	})

	t.Run("DropColumns", func(t *testing.T) {
		mockedDB.Reset()
		if err := engine.DropColumns(model, "active", "updated"); err != nil {
			t.Fatal(err)
		}
		if len(mockedDB.queries) != 1 {
			t.Fatalf("expected 1 queries, got %d", len(mockedDB.queries))
		}
		stmt := mockedDB.queries[0].Stmt
		if !strings.HasPrefix(stmt, `ALTER TABLE "users_user" DROP COLUMN`) {
			t.Errorf(
				"expected query start: %s",
				`ALTER TABLE "users_user" DROP COLUMN`,
			)
		}
	})

	t.Run("SelectQuery", func(t *testing.T) {
		mockedDB.Reset()
		cond := Q{"active": true}.OrNot(
			Q{"email": "user@test.com"}.Or(Q{"id >=": 10}),
		).AndNot(
			Q{"updated <": "2018-07-20"},
		)
		query, err := engine.SelectQuery(model, cond, "id", "email")
		if err != nil {
			t.Fatal(err)
		}
		expected := `SELECT "id", "email" FROM "users_user" WHERE (` +
			`("active" = $1) OR NOT (("email" = $2) OR ("id" >= $3))` +
			`) AND NOT ("updated" < $4)`
		if query.Stmt != expected {
			t.Fatalf("expected:\n\n%s\n\ngot:\n\n%s", expected, query.Stmt)
		}
		if len(query.Args) != 4 {
			t.Fatalf("expected 4 query args, got %d", len(query.Args))
		}
		if val, ok := query.Args[0].(bool); !ok || !val {
			t.Errorf("expected true, got %s", query.Args[0])
		}
		if val, ok := query.Args[3].(string); !ok || val != "2018-07-20" {
			t.Errorf("expected true, got %s", query.Args[3])
		}
	})

	t.Run("GetRows", func(t *testing.T) {
		mockedDB.Reset()
		cond := Q{"active": true}
		_, err := engine.GetRows(model, cond, 10, 20, "id", "updated")
		if err != nil {
			t.Fatal(err)
		}
		if len(mockedDB.queries) != 1 {
			t.Fatalf("expected one query, got %d", len(mockedDB.queries))
		}
		expected := `SELECT "id", "updated" FROM "users_user" ` +
			`WHERE "active" = $1 LIMIT 10 OFFSET 10`
		stmt := mockedDB.queries[0].Stmt
		if stmt != expected {
			t.Fatalf("expected:\n\n%s\n\ngot:\n\n%s", expected, stmt)
		}
	})

	t.Run("InsertRow", func(t *testing.T) {
		mockedDB.Reset()
		values := Values{"email": "user@test.com", "active": true}
		_, err := engine.InsertRow(model, values)
		if err != nil {
			t.Fatal(err)
		}
		if len(mockedDB.queries) != 1 {
			t.Fatalf("expected one query, got %d", len(mockedDB.queries))
		}
		stmt := mockedDB.queries[0].Stmt
		if !strings.HasPrefix(stmt, `INSERT INTO "users_user"`) {
			t.Errorf("expected query start: %s", `INSERT INTO "users_user"`)
		}
		args := mockedDB.queries[0].Args
		if len(args) != 2 {
			t.Fatalf("expected 2 query args, got %d", len(args))
		}
	})

	t.Run("UpdateRows", func(t *testing.T) {
		mockedDB.Reset()
		values := Values{"active": false}
		_, err := engine.UpdateRows(model, values, Q{"active": true})
		if err != nil {
			t.Fatal(err)
		}
		if len(mockedDB.queries) != 1 {
			t.Fatalf("expected one query, got %d", len(mockedDB.queries))
		}
		expected := `UPDATE "users_user" SET "active" = $1 WHERE "active" = $2`
		stmt := mockedDB.queries[0].Stmt
		if stmt != expected {
			t.Fatalf("expected:\n\n%s\n\ngot:\n\n%s", expected, stmt)
		}
		args := mockedDB.queries[0].Args
		if len(args) != 2 {
			t.Fatalf("expected one query args, got %d", len(args))
		}
		if val, ok := args[0].(bool); !ok || val {
			t.Errorf("expected false, got %s", args[0])
		}
	})

	t.Run("DeleteRows", func(t *testing.T) {
		mockedDB.Reset()
		_, err := engine.DeleteRows(model, Q{"id >=": 100})
		if err != nil {
			t.Fatal(err)
		}
		if len(mockedDB.queries) != 1 {
			t.Fatalf("expected one query, got %d", len(mockedDB.queries))
		}
		expected := `DELETE FROM "users_user" WHERE "id" >= $1`
		stmt := mockedDB.queries[0].Stmt
		if stmt != expected {
			t.Fatalf("expected:\n\n%s\n\ngot:\n\n%s", expected, stmt)
		}
		args := mockedDB.queries[0].Args
		if len(args) != 1 {
			t.Fatalf("expected one query args, got %d", len(args))
		}
		if val, ok := args[0].(int); !ok || val != 100 {
			t.Errorf("expected 100, got %s", args[0])
		}
	})

	t.Run("CountRows", func(t *testing.T) {
		mockedDB.Reset()
		_, err := engine.CountRows(model, Q{"email": "user@test.com"})
		if err != nil {
			t.Fatal(err)
		}
		if len(mockedDB.queries) != 1 {
			t.Fatalf("expected one query, got %d", len(mockedDB.queries))
		}
		expected := `SELECT COUNT(*) FROM "users_user" WHERE "email" = $1`
		stmt := mockedDB.queries[0].Stmt
		if stmt != expected {
			t.Fatalf("expected:\n\n%s\n\ngot:\n\n%s", expected, stmt)
		}
		args := mockedDB.queries[0].Args
		if len(args) != 1 {
			t.Fatalf("expected one query args, got %d", len(args))
		}
		if val, ok := args[0].(string); !ok || val != "user@test.com" {
			t.Errorf("expected user@test.com, got %s", args[0])
		}
	})

	t.Run("Exists", func(t *testing.T) {
		mockedDB.Reset()
		_, err := engine.Exists(model, Q{"email": "user@test.com"})
		if err != nil {
			t.Fatal(err)
		}
		if len(mockedDB.queries) != 1 {
			t.Fatalf("expected one query, got %d", len(mockedDB.queries))
		}
		expected := `SELECT EXISTS (` +
			`SELECT "id" FROM "users_user" WHERE "email" = $1)`
		stmt := mockedDB.queries[0].Stmt
		if stmt != expected {
			t.Fatalf("expected:\n\n%s\n\ngot:\n\n%s", expected, stmt)
		}
		args := mockedDB.queries[0].Args
		if len(args) != 1 {
			t.Fatalf("expected one query args, got %d", len(args))
		}
		if val, ok := args[0].(string); !ok || val != "user@test.com" {
			t.Errorf("expected user@test.com, got %s", args[0])
		}
	})
}
