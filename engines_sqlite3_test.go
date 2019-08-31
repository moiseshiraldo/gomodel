package gomodels

import (
	"database/sql"
	"strings"
	"testing"
)

type dbMocker struct {
	queries []Query
}

func (db *dbMocker) Reset() {
	db.queries = make([]Query, 0)
}

func (db *dbMocker) Begin() (*sql.Tx, error) {
	return nil, nil
}

func (db *dbMocker) Close() error {
	return nil
}

func (db *dbMocker) Commit() error {
	return nil
}
func (db *dbMocker) Rollback() error {
	return nil
}

func (db *dbMocker) Exec(stmt string, args ...interface{}) (sql.Result, error) {
	db.queries = append(db.queries, Query{stmt, args})
	return nil, nil
}

func (db *dbMocker) Query(stmt string, args ...interface{}) (*sql.Rows, error) {
	db.queries = append(db.queries, Query{stmt, args})
	return nil, nil
}

func (db *dbMocker) QueryRow(stmt string, args ...interface{}) *sql.Row {
	db.queries = append(db.queries, Query{stmt, args})
	return nil
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
}
