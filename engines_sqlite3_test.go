package gomodels

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
)

type dbMocker struct {
	queries []Query
	err     error
}

func (db *dbMocker) Reset() {
	db.queries = make([]Query, 0)
	db.err = nil
}

func (db *dbMocker) Begin() (*sql.Tx, error) {
	return nil, db.err
}

func (db *dbMocker) Close() error {
	return db.err
}

func (db *dbMocker) Commit() error {
	return db.err
}
func (db *dbMocker) Rollback() error {
	return db.err
}

func (db *dbMocker) Exec(stmt string, args ...interface{}) (sql.Result, error) {
	db.queries = append(db.queries, Query{stmt, args})
	return resultMocker{}, db.err
}

func (db *dbMocker) Query(stmt string, args ...interface{}) (*sql.Rows, error) {
	db.queries = append(db.queries, Query{stmt, args})
	return nil, db.err
}

func (db *dbMocker) QueryRow(stmt string, args ...interface{}) *sql.Row {
	db.queries = append(db.queries, Query{stmt, args})
	return &sql.Row{}
}

type resultMocker struct{}

func (res resultMocker) LastInsertId() (int64, error) {
	return 42, nil
}

func (res resultMocker) RowsAffected() (int64, error) {
	return 1, nil
}

// TestSqliteEngine tests the SqliteEngine methods
func TestSqliteEngine(t *testing.T) {
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
	engine := SqliteEngine{baseSQLEngine{
		db:          mockedDB,
		driver:      "sqlite3",
		escapeChar:  "\"",
		placeholder: "?",
	}}
	origScanRow := scanRow
	origOpenDB := openDB
	defer func() {
		scanRow = origScanRow
		openDB = origOpenDB
	}()
	scanRow = func(ex sqlExecutor, dest interface{}, query Query) error {
		db := ex.(*dbMocker)
		db.queries = append(db.queries, query)
		return nil
	}

	t.Run("StartErr", func(t *testing.T) {
		openDB = func(driver string, credentials string) (*sql.DB, error) {
			return nil, fmt.Errorf("db error")
		}
		engine := SqliteEngine{}
		if _, err := engine.Start(Database{}); err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Start", func(t *testing.T) {
		openDB = func(driver string, credentials string) (*sql.DB, error) {
			return nil, nil
		}
		engine, err := SqliteEngine{}.Start(Database{})
		if err != nil {
			t.Fatal(err)
		}
		eng := engine.(SqliteEngine)
		if eng.baseSQLEngine.driver != "sqlite3" {
			t.Errorf("expected sqlite3, got %s", eng.baseSQLEngine.driver)
		}
		if eng.baseSQLEngine.escapeChar != "\"" {
			t.Errorf("expected \", got %s", eng.baseSQLEngine.escapeChar)
		}
		if eng.baseSQLEngine.placeholder != "?" {
			t.Errorf("expected ?, got %s", eng.baseSQLEngine.placeholder)
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

	t.Run("TxSupport", func(t *testing.T) {
		mockedDB.Reset()
		if !engine.TxSupport() {
			t.Fatalf("expected true, got false")
		}
	})

	t.Run("StopErr", func(t *testing.T) {
		mockedDB.Reset()
		mockedDB.err = fmt.Errorf("db error")
		if err := engine.Stop(); err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("CommitTxErr", func(t *testing.T) {
		mockedDB.Reset()
		if err := engine.CommitTx(); err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("CommitTxDBErr", func(t *testing.T) {
		mockedDB.Reset()
		mockedDB.err = fmt.Errorf("db error")
		engine.baseSQLEngine.tx = mockedDB
		if err := engine.CommitTx(); err == nil {
			t.Error("expected error, got nil")
		}
		engine.baseSQLEngine.tx = nil
	})

	t.Run("RollbackTxErr", func(t *testing.T) {
		mockedDB.Reset()
		if err := engine.RollbackTx(); err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("RollbackTxDBErr", func(t *testing.T) {
		mockedDB.Reset()
		mockedDB.err = fmt.Errorf("db error")
		engine.baseSQLEngine.tx = mockedDB
		if err := engine.RollbackTx(); err == nil {
			t.Error("expected error, got nil")
		}
		engine.baseSQLEngine.tx = nil
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
		if !strings.Contains(stmt, `"id" INTEGER`) {
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
		if len(mockedDB.queries) != 2 {
			t.Fatalf("expected 2 queries, got %d", len(mockedDB.queries))
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
		if len(mockedDB.queries) != 4 {
			t.Fatalf("expected 4 queries, got %d", len(mockedDB.queries))
		}
		st := mockedDB.queries[0].Stmt
		if !strings.HasPrefix(st, `CREATE TABLE "users_user__new" AS SELECT`) {
			t.Fatalf(
				"expected query start: %s",
				`CREATE TABLE "users_user__new" AS SELECT`,
			)
		}
		st = mockedDB.queries[1].Stmt
		if st != `DROP TABLE "users_user"` {
			t.Fatalf(
				"expected:\n\n%s\n\ngot:\n\n%s", `DROP TABLE "users_user"`, st,
			)
		}
		st = mockedDB.queries[2].Stmt
		expected := `ALTER TABLE "users_user__new" RENAME TO "users_user"`
		if st != expected {
			t.Fatalf("expected:\n\n%s\n\ngot:\n\n%s", expected, st)
		}
		st = mockedDB.queries[3].Stmt
		expected = `CREATE INDEX "test_index" ON "users_user" ("email")`
		if st != expected {
			t.Fatalf("expected:\n\n%s\n\ngot:\n\n%s", expected, st)
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
			`("active" = ?) OR NOT (("email" = ?) OR ("id" >= ?))` +
			`) AND NOT ("updated" < ?)`
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
			`WHERE "active" = ? LIMIT 10 OFFSET 10`
		stmt := mockedDB.queries[0].Stmt
		if stmt != expected {
			t.Fatalf("expected:\n\n%s\n\ngot:\n\n%s", expected, stmt)
		}
	})

	t.Run("GetRowsNoLimit", func(t *testing.T) {
		mockedDB.Reset()
		cond := Q{"active": true}
		_, err := engine.GetRows(model, cond, 10, -1, "id", "updated")
		if err != nil {
			t.Fatal(err)
		}
		if len(mockedDB.queries) != 1 {
			t.Fatalf("expected one query, got %d", len(mockedDB.queries))
		}
		expected := `SELECT "id", "updated" FROM "users_user" ` +
			`WHERE "active" = ? LIMIT -1 OFFSET 10`
		stmt := mockedDB.queries[0].Stmt
		if stmt != expected {
			t.Fatalf("expected:\n\n%s\n\ngot:\n\n%s", expected, stmt)
		}
	})

	t.Run("GetRowsInvalidCondition", func(t *testing.T) {
		mockedDB.Reset()
		cond := Q{"username": "test"}
		_, err := engine.GetRows(model, cond, 10, 20, "id", "updated")
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

	t.Run("InsertRowDBError", func(t *testing.T) {
		mockedDB.Reset()
		mockedDB.err = fmt.Errorf("db error")
		values := Values{"email": "user@test.com", "active": true}
		_, err := engine.InsertRow(model, values)
		if err == nil {
			t.Fatal("expected db error")
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
		expected := `UPDATE "users_user" SET "active" = ? WHERE "active" = ?`
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
		expected := `DELETE FROM "users_user" WHERE "id" >= ?`
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
		expected := `SELECT COUNT(*) FROM "users_user" WHERE "email" = ?`
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

	t.Run("CountInvalidCondition", func(t *testing.T) {
		mockedDB.Reset()
		_, err := engine.CountRows(model, Q{"username": "user@test.com"})
		if err == nil {
			t.Fatal("expected unknown field error")
		}
	})

	t.Run("CountDBError", func(t *testing.T) {
		mockedDB.Reset()
		origScanRow := scanRow
		defer func() { scanRow = origScanRow }()
		scanRow = func(ex sqlExecutor, dest interface{}, query Query) error {
			return fmt.Errorf("db error")
		}
		_, err := engine.CountRows(model, Q{"email": "user@test.com"})
		if err == nil {
			t.Fatal("expected db error")
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
			`SELECT "id" FROM "users_user" WHERE "email" = ?)`
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

	t.Run("ExistsInvalidCondition", func(t *testing.T) {
		mockedDB.Reset()
		_, err := engine.Exists(model, Q{"username": "user@test.com"})
		if err == nil {
			t.Fatal("expected unknown field error")
		}
	})

	t.Run("ExistsDBError", func(t *testing.T) {
		mockedDB.Reset()
		origScanRow := scanRow
		defer func() { scanRow = origScanRow }()
		scanRow = func(ex sqlExecutor, dest interface{}, query Query) error {
			return fmt.Errorf("db error")
		}
		_, err := engine.Exists(model, Q{"email": "user@test.com"})
		if err == nil {
			t.Fatal("expected db error")
		}
	})
}
