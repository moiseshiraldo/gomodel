package gomodel

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
)

// TestPostgresEngine tests the PostgresEngine methods
func TestPostgresEngine(t *testing.T) {
	model := &Model{
		name: "User",
		pk:   "id",
		fields: Fields{
			"id":      IntegerField{Auto: true, PrimaryKey: true},
			"email":   CharField{MaxLength: 100, Unique: true},
			"active":  BooleanField{DefaultFalse: true},
			"updated": DateTimeField{AutoNow: true, Null: true},
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

	t.Run("StartErr", func(t *testing.T) {
		openDB = func(driver string, credentials string) (*sql.DB, error) {
			return nil, fmt.Errorf("db error")
		}
		engine := PostgresEngine{}
		if _, err := engine.Start(Database{}); err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Start", func(t *testing.T) {
		openDB = func(driver string, credentials string) (*sql.DB, error) {
			return nil, nil
		}
		engine, err := PostgresEngine{}.Start(Database{})
		if err != nil {
			t.Fatal(err)
		}
		eng := engine.(PostgresEngine)
		if eng.baseSQLEngine.driver != "postgres" {
			t.Errorf("expected postgres, got %s", eng.baseSQLEngine.driver)
		}
		if eng.baseSQLEngine.escapeChar != "\"" {
			t.Errorf("expected \", got %s", eng.baseSQLEngine.escapeChar)
		}
		if eng.baseSQLEngine.placeholder != "$" {
			t.Errorf("expected $, got %s", eng.baseSQLEngine.placeholder)
		}
	})

	t.Run("BeginTxErr", func(t *testing.T) {
		mockedDB.Reset()
		mockedDB.err = fmt.Errorf("db error")
		if _, err := engine.BeginTx(); err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("BeginTx", func(t *testing.T) {
		mockedDB.Reset()
		if _, err := engine.BeginTx(); err != nil {
			t.Fatal(err)
		}
	})

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
			Q{"updated": nil},
		)
		options := QueryOptions{
			Conditioner: cond,
			Fields:      []string{"id", "email"},
		}
		query, err := engine.SelectQuery(model, options)
		if err != nil {
			t.Fatal(err)
		}
		expected := `SELECT "id", "email" FROM "users_user" WHERE (` +
			`("active" = $1) OR NOT (("email" = $2) OR ("id" >= $3))` +
			`) AND NOT ("updated" IS NULL)`
		if query.Stmt != expected {
			t.Fatalf("expected:\n\n%s\n\ngot:\n\n%s", expected, query.Stmt)
		}
		if len(query.Args) != 3 {
			t.Fatalf("expected 4 query args, got %d", len(query.Args))
		}
		if val, ok := query.Args[0].(bool); !ok || !val {
			t.Errorf("expected true, got %s", query.Args[0])
		}
	})

	t.Run("SelectInvalidOperator", func(t *testing.T) {
		mockedDB.Reset()
		cond := Q{"active": true}.OrNot(
			Q{"email": "user@test.com"}.Or(Q{"id +-": 10}),
		).AndNot(
			Q{"updated <": "2018-07-20"},
		)
		options := QueryOptions{
			Conditioner: cond,
			Fields:      []string{"id", "email"},
		}
		_, err := engine.SelectQuery(model, options)
		if err == nil {
			t.Fatal("expected invalid operator error")
		}
	})

	t.Run("SelectUnknownField", func(t *testing.T) {
		mockedDB.Reset()
		cond := Q{"active": true}
		options := QueryOptions{
			Conditioner: cond,
			Fields:      []string{"id", "username"},
		}
		_, err := engine.SelectQuery(model, options)
		if err == nil {
			t.Fatal("expected unknown field error")
		}
	})

	t.Run("SelectUnknownConditionField", func(t *testing.T) {
		mockedDB.Reset()
		cond := Q{"admin": true}.OrNot(
			Q{"email": "user@test.com"}.Or(Q{"id >=": 10}),
		).AndNot(
			Q{"updated <": "2018-07-20"},
		)
		options := QueryOptions{
			Conditioner: cond,
			Fields:      []string{"id", "email"},
		}
		_, err := engine.SelectQuery(model, options)
		if err == nil {
			t.Fatal("expected unknown field error")
		}
	})

	t.Run("SelectInvalidValue", func(t *testing.T) {
		mockedDB.Reset()
		cond := Q{"active": true}.OrNot(
			Q{"email": "user@test.com"}.Or(Q{"id >=": 10}),
		).AndNot(
			Q{"updated <": true},
		)
		options := QueryOptions{
			Conditioner: cond,
			Fields:      []string{"id", "email"},
		}
		_, err := engine.SelectQuery(model, options)
		if err == nil {
			t.Fatal("expected invalid value error")
		}
	})

	t.Run("GetRows", func(t *testing.T) {
		mockedDB.Reset()
		options := QueryOptions{
			Conditioner: Q{"active": true},
			Fields:      []string{"id", "updated"},
			Start:       10,
			End:         20,
		}
		_, err := engine.GetRows(model, options)
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

	t.Run("GetRowsNoLimit", func(t *testing.T) {
		mockedDB.Reset()
		options := QueryOptions{
			Conditioner: Q{"active": true},
			Fields:      []string{"id", "updated"},
			Start:       10,
			End:         -1,
		}
		_, err := engine.GetRows(model, options)
		if err != nil {
			t.Fatal(err)
		}
		if len(mockedDB.queries) != 1 {
			t.Fatalf("expected one query, got %d", len(mockedDB.queries))
		}
		expected := `SELECT "id", "updated" FROM "users_user" ` +
			`WHERE "active" = $1 LIMIT ALL OFFSET 10`
		stmt := mockedDB.queries[0].Stmt
		if stmt != expected {
			t.Fatalf("expected:\n\n%s\n\ngot:\n\n%s", expected, stmt)
		}
	})

	t.Run("GetRowsInvalidCondition", func(t *testing.T) {
		mockedDB.Reset()
		options := QueryOptions{
			Conditioner: Q{"username": "test"},
			Fields:      []string{"id", "updated"},
			Start:       10,
			End:         20,
		}
		_, err := engine.GetRows(model, options)
		if err == nil {
			t.Fatal("expected unknown field error")
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

	t.Run("InsertUnknownField", func(t *testing.T) {
		mockedDB.Reset()
		values := Values{"username": "test", "active": true}
		_, err := engine.InsertRow(model, values)
		if err == nil {
			t.Fatal("expected unknown field error")
		}
	})

	t.Run("InsertInvalidValue", func(t *testing.T) {
		mockedDB.Reset()
		values := Values{"email": "user@test.com", "updated": true}
		_, err := engine.InsertRow(model, values)
		if err == nil {
			t.Fatal("expected invalid value error")
		}
	})

	t.Run("InsertRowDBError", func(t *testing.T) {
		mockedDB.Reset()
		origScanRow := scanRow
		defer func() { scanRow = origScanRow }()
		scanRow = func(ex sqlExecutor, dest interface{}, query Query) error {
			return fmt.Errorf("db error")
		}
		values := Values{"email": "user@test.com", "active": true}
		_, err := engine.InsertRow(model, values)
		if err == nil {
			t.Fatal("expected db error")
		}
	})

	t.Run("UpdateRows", func(t *testing.T) {
		mockedDB.Reset()
		values := Values{"active": false}
		options := QueryOptions{Conditioner: Q{"active": true}}
		_, err := engine.UpdateRows(model, values, options)
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

	t.Run("UpdateUnknownField", func(t *testing.T) {
		mockedDB.Reset()
		values := Values{"username": "test"}
		options := QueryOptions{Conditioner: Q{"active": true}}
		_, err := engine.UpdateRows(model, values, options)
		if err == nil {
			t.Fatal("expected unknown field error")
		}
	})

	t.Run("UpdateUnknownConditionField", func(t *testing.T) {
		mockedDB.Reset()
		values := Values{"active": false}
		options := QueryOptions{Conditioner: Q{"username": "test"}}
		_, err := engine.UpdateRows(model, values, options)
		if err == nil {
			t.Fatal("expected unknown field error")
		}
	})

	t.Run("UpdateInvalidValue", func(t *testing.T) {
		mockedDB.Reset()
		values := Values{"updated": true}
		options := QueryOptions{Conditioner: Q{"active": true}}
		_, err := engine.UpdateRows(model, values, options)
		if err == nil {
			t.Fatal("expected invalid value error")
		}
	})

	t.Run("UpdateDBError", func(t *testing.T) {
		mockedDB.Reset()
		mockedDB.err = fmt.Errorf("db error")
		values := Values{"active": false}
		options := QueryOptions{Conditioner: Q{"active": true}}
		_, err := engine.UpdateRows(model, values, options)
		if err == nil {
			t.Fatal("expected db error")
		}
	})

	t.Run("DeleteRows", func(t *testing.T) {
		mockedDB.Reset()
		options := QueryOptions{Conditioner: Q{"id >=": 100}}
		_, err := engine.DeleteRows(model, options)
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

	t.Run("DeleteInvalidCondition", func(t *testing.T) {
		mockedDB.Reset()
		options := QueryOptions{Conditioner: Q{"loginAttempts >=": 100}}
		_, err := engine.DeleteRows(model, options)
		if err == nil {
			t.Fatal("expected unknown field error")
		}
	})

	t.Run("DeleteDBError", func(t *testing.T) {
		mockedDB.Reset()
		mockedDB.err = fmt.Errorf("db error")
		options := QueryOptions{Conditioner: Q{"id >=": 100}}
		_, err := engine.DeleteRows(model, options)
		if err == nil {
			t.Fatal("expected unknown field error")
		}
	})

	t.Run("CountRows", func(t *testing.T) {
		mockedDB.Reset()
		options := QueryOptions{Conditioner: Q{"email": "user@test.com"}}
		_, err := engine.CountRows(model, options)
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
		engine.baseSQLEngine.tx = mockedDB
		options := QueryOptions{Conditioner: Q{"email": "user@test.com"}}
		_, err := engine.Exists(model, options)
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
