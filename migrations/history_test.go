package migrations

import (
	"fmt"
	"github.com/moiseshiraldo/gomodels"
	"strings"
	"testing"
)

// rowsMocker mocks the results returned by the database engine
type rowsMocker struct {
	finished bool
}

func (r *rowsMocker) Close() error {
	return nil
}

func (r rowsMocker) Err() error {
	return nil
}

func (r *rowsMocker) Next() bool {
	if r.finished {
		return false
	}
	r.finished = true
	return true
}

func (r *rowsMocker) Scan(dest ...interface{}) error {
	appName := dest[0].(*string)
	number := dest[1].(*int)
	*appName = "test"
	*number = 1
	return nil
}

// TestAppMigrate tests the Migrate method of the app state
func TestAppMigrate(t *testing.T) {
	// App setup
	gomodels.Register(gomodels.NewApp("test", ""))
	defer gomodels.ClearRegistry()
	// DB setup
	dbSettings := map[string]gomodels.Database{
		"default": {Driver: "mocker", Name: "test"},
	}
	if err := gomodels.Start(dbSettings); err != nil {
		t.Fatal(err)
	}
	db := gomodels.Databases()["default"]
	mockedEngine := db.Engine.(gomodels.MockedEngine)
	defer gomodels.Stop()
	// App state setup
	firstNode := &Node{App: "test", Name: "initial", number: 1}
	secondNode := &Node{
		App:          "test",
		Name:         "test_migrations",
		number:       2,
		Dependencies: [][]string{{"test", "0001_initial"}},
	}
	appState := &AppState{
		app: gomodels.Registry()["test"],
	}
	history["test"] = appState
	defer clearHistory()

	t.Run("NoMigrations", func(t *testing.T) {
		err := appState.Migrate("default", "")
		if _, ok := err.(*NoAppMigrationsError); !ok {
			t.Errorf("expected NoAppMigrationsError, got %T", err)
		}
	})

	appState.migrations = []*Node{firstNode, secondNode}

	t.Run("NoDatabase", func(t *testing.T) {
		err := appState.Migrate("Slave", "")
		if _, ok := err.(*gomodels.DatabaseError); !ok {
			t.Errorf("expected gomodels.DatabaseError, got %T", err)
		}
	})

	t.Run("InvalidNodeName", func(t *testing.T) {
		err := appState.Migrate("default", "TestName")
		if _, ok := err.(*NameError); !ok {
			t.Errorf("expected NameError, got %T", err)
		}
	})

	t.Run("InvalidNodeNumber", func(t *testing.T) {
		err := appState.Migrate("default", "0003_test_migration")
		if _, ok := err.(*NameError); !ok {
			t.Errorf("expected NameError, got %T", err)
		}
	})

	t.Run("MigrateFirstNode", func(t *testing.T) {
		mockedEngine.Reset()
		if err := appState.Migrate("default", "0001"); err != nil {
			t.Fatal(err)
		}
		if !appState.migrations[0].applied {
			t.Errorf("first migration was not applied")
		}
		if appState.migrations[1].applied {
			t.Errorf("second migration was applied")
		}
		if mockedEngine.Calls("SaveMigration") != 1 {
			t.Errorf("expected engine SaveMigration to be called")
		}
		args := mockedEngine.Args.SaveMigration
		if args.App != "test" || args.Number != 1 {
			t.Errorf(
				"SaveMigration called with wrong arguments: %s, %d",
				args.App, args.Number,
			)
		}
	})

	t.Run("MigrateAll", func(t *testing.T) {
		firstNode.applied = false
		secondNode.applied = false
		mockedEngine.Reset()
		if err := appState.Migrate("default", ""); err != nil {
			t.Fatal(err)
		}
		if !appState.migrations[0].applied {
			t.Errorf("first migration was not applied")
		}
		if !appState.migrations[1].applied {
			t.Errorf("second migration was not applied")
		}
		if mockedEngine.Calls("SaveMigration") != 2 {
			t.Errorf("expected engine SaveMigration to be called twice")
		}
		args := mockedEngine.Args.SaveMigration
		if args.App != "test" || args.Number != 2 {
			t.Errorf(
				"SaveMigration called with wrong arguments: %s, %d",
				args.App, args.Number,
			)
		}
	})

	t.Run("MigrateBackwardsFirst", func(t *testing.T) {
		firstNode.applied = true
		secondNode.applied = true
		appState.lastApplied = 2
		mockedEngine.Reset()
		if err := appState.Migrate("default", "0001"); err != nil {
			t.Fatal(err)
		}
		if !appState.migrations[0].applied {
			t.Errorf("first migration is not applied")
		}
		if appState.migrations[1].applied {
			t.Errorf("second migration is still applied")
		}
		if mockedEngine.Calls("DeleteMigration") != 1 {
			t.Errorf("expected engine DeleteMigration to be called")
		}
		args := mockedEngine.Args.DeleteMigration
		if args.App != "test" || args.Number != 2 {
			t.Errorf(
				"SaveMigration called with wrong arguments: %s, %d",
				args.App, args.Number,
			)
		}
	})

	t.Run("MigrateBackwardsAll", func(t *testing.T) {
		firstNode.applied = true
		secondNode.applied = true
		appState.lastApplied = 2
		mockedEngine.Reset()
		if err := appState.Migrate("default", "0000"); err != nil {
			t.Fatal(err)
		}
		if appState.migrations[0].applied {
			t.Errorf("first migration is still applied")
		}
		if appState.migrations[1].applied {
			t.Errorf("second migration is still applied")
		}
		if mockedEngine.Calls("DeleteMigration") != 2 {
			t.Errorf("expected engine DeleteMigration to be called twice")
		}
		args := mockedEngine.Args.DeleteMigration
		if args.App != "test" || args.Number != 1 {
			t.Errorf(
				"SaveMigration called with wrong arguments: %s, %d",
				args.App, args.Number,
			)
		}
	})
}

// TestAppMakeMigrations tests the MakeMigrations method of the app state
func TestAppMakeMigrations(t *testing.T) {
	// Models setup
	user := gomodels.New(
		"User",
		gomodels.Fields{
			"email": gomodels.CharField{MaxLength: 100, Index: true},
		},
		gomodels.Options{Table: "users"},
	)
	customer := gomodels.New(
		"Customer",
		gomodels.Fields{
			"name": gomodels.CharField{MaxLength: 100},
		},
		gomodels.Options{
			Indexes: gomodels.Indexes{"initial_idx": []string{"name"}},
		},
	)
	// Apps setup
	usersApp := gomodels.NewApp("users", "", user.Model)
	customersApp := gomodels.NewApp("customers", "", customer.Model)
	gomodels.Register(usersApp, customersApp)
	defer gomodels.ClearRegistry()
	// App states setup
	operation := CreateModel{Name: "Customer", Fields: customer.Model.Fields()}
	node := &Node{
		App:        "customers",
		Name:       "initial",
		number:     1,
		Operations: OperationList{operation},
		processed:  true,
	}
	history["users"] = &AppState{
		app:    gomodels.Registry()["users"],
		Models: make(map[string]*gomodels.Model),
	}
	history["customers"] = &AppState{
		app:        gomodels.Registry()["customers"],
		Models:     map[string]*gomodels.Model{"Customer": customer.Model},
		migrations: []*Node{node},
	}
	defer clearHistory()

	t.Run("NoChanges", func(t *testing.T) {
		migrations, err := history["customers"].MakeMigrations()
		if err != nil {
			t.Fatal(err)
		}
		if len(migrations) > 0 {
			t.Fatalf("expected no migrations, got %d", len(migrations))
		}
	})

	t.Run("Initial", func(t *testing.T) {
		migrations, err := history["users"].MakeMigrations()
		if err != nil {
			t.Fatal(err)
		}
		if len(migrations) != 1 {
			t.Fatalf("expected 1 migration, got %d", len(migrations))
		}
		if migrations[0].number != 1 {
			t.Errorf("expected node number 1, got %d", migrations[0].number)
		}
		if len(migrations[0].Operations) != 2 {
			t.Fatalf("expected migration to contain two operations")
		}
		if migrations[0].Operations[0].OpName() != "CreateModel" {
			name := migrations[0].Operations[0]
			t.Fatalf("expected CreateModel operation, got %s", name)
		}
		createOp := migrations[0].Operations[0].(CreateModel)
		if createOp.Name != "User" || createOp.Table != "users" {
			t.Errorf("operation CreateModel has wrong details")
		}
		if _, ok := createOp.Fields["email"]; !ok {
			t.Errorf("operation CreateModel missing name field")
		}
		if _, ok := history["users"].Models["User"]; !ok {
			t.Fatal("operation CreateModel was not applied to state")
		}
		if migrations[0].Operations[1].OpName() != "AddIndex" {
			name := migrations[0].Operations[1].OpName()
			t.Fatalf("expected AddIndex operation, got %s", name)
		}
		idxOp := migrations[0].Operations[1].(AddIndex)
		if idxOp.Model != "User" || idxOp.Name != "users_user_email_auto_idx" {
			t.Errorf("operation AddIndex has wrong details")
		}
		if len(idxOp.Fields) != 1 && idxOp.Fields[0] != "email" {
			t.Errorf("operation AddIndex missing email field")
		}
		if len(history["users"].Models["User"].Indexes()) == 0 {
			t.Errorf("operation AddIndex was not applied to state")
		}
	})

	t.Run("AddField", func(t *testing.T) {
		customerState := gomodels.New(
			"Customer",
			gomodels.Fields{},
			gomodels.Options{
				Table:   customer.Model.Table(),
				Indexes: customer.Model.Indexes(),
			},
		)
		history["customers"].Models["Customer"] = customerState.Model
		migrations, err := history["customers"].MakeMigrations()
		if err != nil {
			t.Fatal(err)
		}
		if len(migrations) != 1 {
			t.Fatalf("expected 1 migration, got %d", len(migrations))
		}
		if migrations[0].number != 2 {
			t.Errorf("expected node number 2, got %d", migrations[0].number)
		}
		if len(migrations[0].Operations) != 1 {
			t.Fatal("expected migration to contain one operation")
		}
		if migrations[0].Operations[0].OpName() != "AddFields" {
			name := migrations[0].Operations[0].OpName()
			t.Fatalf("expected AddFields operation, got %s", name)
		}
		fieldOp := migrations[0].Operations[0].(AddFields)
		if fieldOp.Model != "Customer" {
			t.Errorf("operation AddFields has wrong model")
		}
		if _, ok := fieldOp.Fields["name"]; !ok {
			t.Errorf("operation AddFields missing name field")
		}
		modelState := history["customers"].Models["Customer"]
		if _, ok := modelState.Fields()["name"]; !ok {
			t.Errorf("operation AddFields was not applied to state")
		}
		history["customers"].Models["Customer"] = customer.Model
		history["customers"].migrations = []*Node{node}
	})

	t.Run("RemoveField", func(t *testing.T) {
		fields := customer.Model.Fields()
		fields["active"] = gomodels.BooleanField{}
		customerState := gomodels.New(
			"Customer",
			fields,
			gomodels.Options{
				Table:   customer.Model.Table(),
				Indexes: customer.Model.Indexes(),
			},
		)
		history["customers"].Models["Customer"] = customerState.Model
		migrations, err := history["customers"].MakeMigrations()
		if err != nil {
			t.Fatal(err)
		}
		if len(migrations) != 1 {
			t.Fatalf("expected 1 migration, got %d", len(migrations))
		}
		if migrations[0].number != 2 {
			t.Errorf("expected node number 2, got %d", migrations[0].number)
		}
		if len(migrations[0].Operations) != 1 {
			t.Fatal("expected migration to contain one operation")
		}
		if migrations[0].Operations[0].OpName() != "RemoveFields" {
			name := migrations[0].Operations[0].OpName()
			t.Fatalf("expected RemoveFields operation, got %s", name)
		}
		fieldOp := migrations[0].Operations[0].(RemoveFields)
		if fieldOp.Model != "Customer" {
			t.Errorf("operation RemoveFields has wrong model")
		}
		if len(fieldOp.Fields) != 1 || fieldOp.Fields[0] != "active" {
			t.Errorf("operation RemoveFields missing active field")
		}
		modelState := history["customers"].Models["Customer"]
		if _, found := modelState.Fields()["active"]; found {
			t.Errorf("operation RemoveFields was not applied to state")
		}
		history["customers"].Models["Customer"] = customer.Model
		history["customers"].migrations = []*Node{node}
	})

	t.Run("AddIndex", func(t *testing.T) {
		customerState := gomodels.New(
			"Customer",
			customer.Model.Fields(),
			gomodels.Options{Table: customer.Model.Table()},
		)
		history["customers"].Models["Customer"] = customerState.Model
		migrations, err := history["customers"].MakeMigrations()
		if err != nil {
			t.Fatal(err)
		}
		if len(migrations) != 1 {
			t.Fatalf("expected 1 migration, got %d", len(migrations))
		}
		if migrations[0].number != 2 {
			t.Errorf("expected node number 2, got %d", migrations[0].number)
		}
		if len(migrations[0].Operations) != 1 {
			t.Fatal("expected migration to contain one operation")
		}
		if migrations[0].Operations[0].OpName() != "AddIndex" {
			name := migrations[0].Operations[0].OpName()
			t.Fatalf("expected AddIndex operation, got %s", name)
		}
		idxOp := migrations[0].Operations[0].(AddIndex)
		if idxOp.Model != "Customer" || idxOp.Name != "initial_idx" {
			t.Errorf("operation AddIndex has wrong details")
		}
		modelState := history["customers"].Models["Customer"]
		if _, ok := modelState.Indexes()["initial_idx"]; !ok {
			t.Errorf("operation AddIndex was not applied to state")
		}
		history["customers"].Models["Customer"] = customer.Model
		history["customers"].migrations = []*Node{node}
	})

	t.Run("RemoveIndex", func(t *testing.T) {
		customerState := gomodels.New(
			"Customer",
			customer.Model.Fields(),
			gomodels.Options{
				Indexes: gomodels.Indexes{
					"initial_idx": []string{"name"},
					"new_idx":     []string{"name"},
				},
				Table: customer.Model.Table(),
			},
		)
		history["customers"].Models["Customer"] = customerState.Model
		migrations, err := history["customers"].MakeMigrations()
		if err != nil {
			t.Fatal(err)
		}
		if len(migrations) != 1 {
			t.Fatalf("expected 1 migration, got %d", len(migrations))
		}
		if migrations[0].number != 2 {
			t.Errorf("expected node number 2, got %d", migrations[0].number)
		}
		if len(migrations[0].Operations) != 1 {
			t.Fatal("expected migration to contain one operation")
		}
		if migrations[0].Operations[0].OpName() != "RemoveIndex" {
			name := migrations[0].Operations[0].OpName()
			t.Fatalf("expected RemoveIndex operation, got %s", name)
		}
		idxOp := migrations[0].Operations[0].(RemoveIndex)
		if idxOp.Model != "Customer" || idxOp.Name != "new_idx" {
			t.Errorf("operation RemoveIndex has wrong details")
		}
		modelState := history["customers"].Models["Customer"]
		if _, found := modelState.Indexes()["new_idx"]; found {
			t.Errorf("operation RemoveIndex was not applied to state")
		}
		history["customers"].Models["Customer"] = customer.Model
		history["customers"].migrations = []*Node{node}
	})

	t.Run("DeleteModel", func(t *testing.T) {
		transaction := gomodels.New(
			"Transaction",
			gomodels.Fields{},
			gomodels.Options{},
		)
		history["customers"].Models["Transaction"] = transaction.Model
		migrations, err := history["customers"].MakeMigrations()
		if err != nil {
			t.Fatal(err)
		}
		if len(migrations) != 1 {
			t.Fatalf("expected 1 migration, got %d", len(migrations))
		}
		if migrations[0].number != 2 {
			t.Errorf("expected node number 2, got %d", migrations[0].number)
		}
		if len(migrations[0].Operations) != 1 {
			t.Fatal("expected migration to contain one operation")
		}
		if migrations[0].Operations[0].OpName() != "DeleteModel" {
			name := migrations[0].Operations[0].OpName()
			t.Fatalf("expected DeleteModel operation, got %s", name)
		}
		deleteOp := migrations[0].Operations[0].(DeleteModel)
		if deleteOp.Name != "Transaction" {
			t.Errorf("operation DeleteModel has wrong details")
		}
		if _, found := history["customers"].Models["Transaction"]; found {
			t.Errorf("operation DeleteModel was not applied to state")
		}
		history["customers"].Models["Customer"] = customer.Model
		history["customers"].migrations = []*Node{node}
	})
}

// TestLoadHistory tests the loadHistory function
func TestLoadHistory(t *testing.T) {
	// App setup
	app := gomodels.NewApp("test", "test/migrations")
	gomodels.Register(app)
	defer gomodels.ClearRegistry()
	// Mocks file read/write functions
	origReadAppNodes := readAppNodes
	origReadNode := readNode
	defer func() {
		readAppNodes = origReadAppNodes
		readNode = origReadNode
	}()
	// Registers mocked operation
	if _, ok := operationsRegistry["MockedOperation"]; !ok {
		operationsRegistry["MockedOperation"] = &mockedOperation{}
	}

	t.Run("WrongPath", func(t *testing.T) {
		readAppNodes = func(path string) ([]string, error) {
			return nil, fmt.Errorf("wrong path")
		}
		err := loadHistory()
		if _, ok := err.(*PathError); !ok {
			t.Errorf("expected PathError, got %T", err)
		}
	})

	t.Run("WrongName", func(t *testing.T) {
		readAppNodes = func(path string) ([]string, error) {
			return []string{"0001_initial.yaml"}, nil
		}
		err := loadHistory()
		if _, ok := err.(*NameError); !ok {
			t.Errorf("expected NameError, got %T", err)
		}
	})

	t.Run("WrongFile", func(t *testing.T) {
		readAppNodes = func(path string) ([]string, error) {
			return []string{"0001_initial.json"}, nil
		}
		readNode = func(path string) ([]byte, error) {
			return []byte("-"), nil
		}
		err := loadHistory()
		if _, ok := err.(*LoadError); !ok {
			t.Errorf("expected LoadError, got %T", err)
		}
	})

	t.Run("Duplicate", func(t *testing.T) {
		readAppNodes = func(path string) ([]string, error) {
			return []string{"0001_initial.json", "0001_migration.json"}, nil
		}
		readNode = func(path string) ([]byte, error) {
			return []byte(`{"App": "test", "Dependencies": []}`), nil
		}
		err := loadHistory()
		if _, ok := err.(*DuplicateNumberError); !ok {
			t.Errorf("expected DuplicateNumberError, got %T", err)
		}
	})

	t.Run("WrongDependencyName", func(t *testing.T) {
		readAppNodes = func(path string) ([]string, error) {
			return []string{"0001_initial.json"}, nil
		}
		readNode = func(path string) ([]byte, error) {
			data := `{"App": "test", "Dependencies": [["test", "qwerty"]]}`
			return []byte(data), nil
		}
		err := loadHistory()
		if _, ok := err.(*InvalidDependencyError); !ok {
			t.Errorf("expected InvalidDependencyError, got %T", err)
		}
	})

	t.Run("UnknownDependencyApp", func(t *testing.T) {
		readAppNodes = func(path string) ([]string, error) {
			return []string{"0001_initial.json"}, nil
		}
		readNode = func(path string) ([]byte, error) {
			data := `{"App": "test", "Dependencies": [["users", "0001_a"]]}`
			return []byte(data), nil
		}
		err := loadHistory()
		if _, ok := err.(*InvalidDependencyError); !ok {
			t.Errorf("expected InvalidDependencyError, got %T", err)
		}
	})

	t.Run("WrongDependencyNumber", func(t *testing.T) {
		readAppNodes = func(path string) ([]string, error) {
			return []string{"0001_initial.json"}, nil
		}
		readNode = func(path string) ([]byte, error) {
			data := `{"App": "test", "Dependencies": [["test", "0004_node"]]}`
			return []byte(data), nil
		}
		err := loadHistory()
		if _, ok := err.(*InvalidDependencyError); !ok {
			t.Errorf("expected InvalidDependencyError, got %T", err)
		}
	})

	t.Run("DifferentDependencyName", func(t *testing.T) {
		readAppNodes = func(path string) ([]string, error) {
			return []string{"0001_initial.json"}, nil
		}
		readNode = func(path string) ([]byte, error) {
			data := `{"App": "test", "Dependencies": [["test", "0001_a"]]}`
			return []byte(data), nil
		}
		err := loadHistory()
		if _, ok := err.(*InvalidDependencyError); !ok {
			t.Errorf("expected InvalidDependencyError, got %T", err)
		}
	})

	t.Run("CircularDependency", func(t *testing.T) {
		readAppNodes = func(path string) ([]string, error) {
			return []string{"0001_a.json", "0002_b.json"}, nil
		}
		readNode = func(path string) ([]byte, error) {
			data := ""
			if strings.HasSuffix(path, "0001_a.json") {
				data = `{"App": "test", "Dependencies": [["test", "0002_b"]]}`
			} else {
				data = `{"App": "test", "Dependencies": [["test", "0001_a"]]}`
			}
			return []byte(data), nil
		}
		err := loadHistory()
		if _, ok := err.(*CircularDependencyError); !ok {
			t.Errorf("expected CircularDependencyError, got %T", err)
		}
	})

	t.Run("OperationError", func(t *testing.T) {
		readAppNodes = func(path string) ([]string, error) {
			return []string{"0001_initial.json"}, nil
		}
		readNode = func(path string) ([]byte, error) {
			data := []byte(`{
			  "App": "test",
			  "Dependencies": [],
			  "Operations": [{"MockedOperation": {"StateErr": true}}]
			}`)
			return []byte(data), nil
		}
		err := loadHistory()
		if _, ok := err.(*OperationStateError); !ok {
			t.Errorf("expected OperationStateError, got %T", err)
		}
	})

	t.Run("Success", func(t *testing.T) {
		readAppNodes = func(path string) ([]string, error) {
			return []string{"0001_initial.json"}, nil
		}
		readNode = func(path string) ([]byte, error) {
			return []byte(`{"App": "test", "Dependencies": []}`), nil
		}
		if err := loadHistory(); err != nil {
			t.Fatal(err)
		}
		migrations := history["test"].migrations
		if len(migrations) != 1 {
			t.Fatalf("expected one loaded migration, got %d", len(migrations))
		}
		if !migrations[0].processed {
			t.Errorf("expected node state to be processed")
		}
	})
}

// TestLoadAppliedMigrations test the loadAppliedMigrations function
func TestLoadAppliedMigrations(t *testing.T) {
	// App setup
	app := gomodels.NewApp("test", "")
	gomodels.Register(app)
	defer gomodels.ClearRegistry()
	// App state setup
	appState := &AppState{
		app:        gomodels.Registry()["test"],
		migrations: []*Node{},
	}
	history["test"] = appState
	defer clearHistory()
	// DB Setup
	err := gomodels.Start(map[string]gomodels.Database{
		"default": {Driver: "mocker", Name: "test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	db := gomodels.Databases()["default"]
	defer gomodels.Stop()
	mockedEngine := db.Engine.(gomodels.MockedEngine)

	t.Run("Error", func(t *testing.T) {
		mockedEngine.Results.GetMigrations.Rows = &rowsMocker{}
		if err := loadAppliedMigrations(db); err == nil {
			t.Fatal("expected missing node error, got nil")
		}
		if mockedEngine.Calls("PrepareMigrations") != 1 {
			t.Fatal("expected engine PrepareMigrations to be called")
		}
		if mockedEngine.Calls("GetMigrations") != 1 {
			t.Fatal("expected engine GetMigrations to be called")
		}
		mockedEngine.Reset()
	})

	t.Run("Success", func(t *testing.T) {
		node := &Node{
			App:    "test",
			Name:   "initial",
			number: 1,
		}
		history["test"].migrations = []*Node{node}
		mockedEngine.Results.GetMigrations.Rows = &rowsMocker{}
		if err := loadAppliedMigrations(db); err != nil {
			t.Fatal(err)
		}
		if !history["test"].migrations[0].applied {
			t.Fatal("expected migration to be applied")
		}
		mockedEngine.Reset()
	})
}
