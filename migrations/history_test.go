package migrations

import (
	"fmt"
	"github.com/moiseshiraldo/gomodels"
	"testing"
)

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

func TestAppMigrate(t *testing.T) {
	if err := gomodels.Register(gomodels.NewApp("test", "")); err != nil {
		t.Fatal(err)
	}
	err := gomodels.Start(gomodels.DBSettings{
		"default": {Driver: "mocker", Name: "test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer gomodels.Stop()
	defer gomodels.ClearRegistry()
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
			t.Errorf("Expected NoAppMigrationsError, got %T", err)
		}
	})
	appState.migrations = []*Node{firstNode, secondNode}
	t.Run("NoDatabase", func(t *testing.T) {
		err := appState.Migrate("SlaveDB", "")
		if _, ok := err.(*gomodels.DatabaseError); !ok {
			t.Errorf("Expected gomodels.DatabaseError, got %T", err)
		}
	})
	t.Run("InvalidNodeName", func(t *testing.T) {
		err := appState.Migrate("default", "TestName")
		if _, ok := err.(*NameError); !ok {
			t.Errorf("Expected NameError, got %T", err)
		}
	})
	t.Run("InvalidNodeNumber", func(t *testing.T) {
		err := appState.Migrate("default", "0003_test_migration")
		if _, ok := err.(*NameError); !ok {
			t.Errorf("Expected NameError, got %T", err)
		}
	})
	t.Run("MigrateFirstNode", func(t *testing.T) {
		if err := appState.Migrate("default", "0001"); err != nil {
			t.Fatal(err)
		}
		if !appState.migrations[0].applied {
			t.Errorf("First migration was not applied")
		}
		if appState.migrations[1].applied {
			t.Errorf("Second migration was applied")
		}
	})
	t.Run("MigrateAll", func(t *testing.T) {
		firstNode.applied = false
		secondNode.applied = false
		if err := appState.Migrate("default", ""); err != nil {
			t.Fatal(err)
		}
		if !appState.migrations[0].applied {
			t.Errorf("First migration was not applied")
		}
		if !appState.migrations[1].applied {
			t.Errorf("Second migration was not applied")
		}
	})
	t.Run("MigrateBackwardsFirst", func(t *testing.T) {
		firstNode.applied = true
		secondNode.applied = true
		appState.lastApplied = 2
		if err := appState.Migrate("default", "0001"); err != nil {
			t.Fatal(err)
		}
		if !appState.migrations[0].applied {
			t.Errorf("First migration is not applied")
		}
		if appState.migrations[1].applied {
			t.Errorf("Second migration is still applied")
		}
	})
	t.Run("MigrateBackwardsAll", func(t *testing.T) {
		firstNode.applied = true
		secondNode.applied = true
		appState.lastApplied = 2
		if err := appState.Migrate("default", "0000"); err != nil {
			t.Fatal(err)
		}
		if appState.migrations[0].applied {
			t.Errorf("First migration is still applied")
		}
		if appState.migrations[1].applied {
			t.Errorf("Second migration is still applied")
		}
	})
}

func TestAppMakeMigrations(t *testing.T) {
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
	usersApp := gomodels.NewApp("users", "", user.Model)
	customersApp := gomodels.NewApp("customers", "", customer.Model)
	if err := gomodels.Register(usersApp, customersApp); err != nil {
		t.Fatal(err)
	}
	defer gomodels.ClearRegistry()
	operation := &CreateModel{Name: "Customer", Fields: customer.Model.Fields()}
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
			t.Fatal("expected no migrations")
		}
	})
	t.Run("Initial", func(t *testing.T) {
		migrations, err := history["users"].MakeMigrations()
		if err != nil {
			t.Fatal(err)
		}
		if len(migrations) != 1 {
			t.Fatal("expected one created node")
		}
		if migrations[0].number != 1 {
			t.Errorf("expected node number 1, got %d", migrations[0].number)
		}
		if len(migrations[0].Operations) != 2 {
			t.Fatal("expected migration to contain two operations")
		}
		if migrations[0].Operations[0].OpName() != "CreateModel" {
			name := migrations[0].Operations[0]
			t.Fatalf("expected CreateModel, got %s", name)
		}
		createOp := migrations[0].Operations[0].(*CreateModel)
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
			t.Fatalf("expected AddIndex, got %s", name)
		}
		idxOp := migrations[0].Operations[1].(*AddIndex)
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
			t.Fatal("expected one created node")
		}
		if migrations[0].number != 2 {
			t.Errorf("expected node number 2, got %d", migrations[0].number)
		}
		if len(migrations[0].Operations) != 1 {
			t.Fatal("expected migration to contain one operation")
		}
		if migrations[0].Operations[0].OpName() != "AddFields" {
			name := migrations[0].Operations[0].OpName()
			t.Fatalf("expected AddFields, got %s", name)
		}
		fieldOp := migrations[0].Operations[0].(*AddFields)
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
			t.Fatal("expected one created node")
		}
		if migrations[0].number != 2 {
			t.Errorf("expected node number 2, got %d", migrations[0].number)
		}
		if len(migrations[0].Operations) != 1 {
			t.Fatal("expected migration to contain one operation")
		}
		if migrations[0].Operations[0].OpName() != "RemoveFields" {
			name := migrations[0].Operations[0].OpName()
			t.Fatalf("expected RemoveFields, got %s", name)
		}
		fieldOp := migrations[0].Operations[0].(*RemoveFields)
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
			t.Fatal("expected one created node")
		}
		if migrations[0].number != 2 {
			t.Errorf("expected node number 2, got %d", migrations[0].number)
		}
		if len(migrations[0].Operations) != 1 {
			t.Fatal("expected migration to contain one operation")
		}
		if migrations[0].Operations[0].OpName() != "AddIndex" {
			name := migrations[0].Operations[0].OpName()
			t.Fatalf("expected AddIndex, got %s", name)
		}
		idxOp := migrations[0].Operations[0].(*AddIndex)
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
			t.Fatal("expected one created node")
		}
		if migrations[0].number != 2 {
			t.Errorf("expected node number 2, got %d", migrations[0].number)
		}
		if len(migrations[0].Operations) != 1 {
			t.Fatal("expected migration to contain one operation")
		}
		if migrations[0].Operations[0].OpName() != "RemoveIndex" {
			name := migrations[0].Operations[0].OpName()
			t.Fatalf("expected RemoveIndex, got %s", name)
		}
		idxOp := migrations[0].Operations[0].(*RemoveIndex)
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
			t.Fatal("expected one created node")
		}
		if migrations[0].number != 2 {
			t.Errorf("expected node number 2, got %d", migrations[0].number)
		}
		if len(migrations[0].Operations) != 1 {
			t.Fatal("expected migration to contain one operation")
		}
		if migrations[0].Operations[0].OpName() != "DeleteModel" {
			name := migrations[0].Operations[0].OpName()
			t.Fatalf("expected DeleteModel, got %s", name)
		}
		deleteOp := migrations[0].Operations[0].(*DeleteModel)
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

func TestLoadHistory(t *testing.T) {
	app := gomodels.NewApp("test", "test/migrations")
	if err := gomodels.Register(app); err != nil {
		t.Fatal(err)
	}
	defer gomodels.ClearRegistry()
	origReadAppNodes := readAppNodes
	origReadNode := readNode
	defer func() {
		readAppNodes = origReadAppNodes
		readNode = origReadNode
	}()
	t.Run("WrongPath", func(t *testing.T) {
		readAppNodes = func(path string) ([]string, error) {
			return nil, fmt.Errorf("wrong path")
		}
		err := loadHistory()
		if _, ok := err.(*PathError); !ok {
			t.Errorf("Expected PathError, got %T", err)
		}
	})
	t.Run("WrongName", func(t *testing.T) {
		readAppNodes = func(path string) ([]string, error) {
			return []string{"0001_initial.yaml"}, nil
		}
		err := loadHistory()
		if _, ok := err.(*NameError); !ok {
			t.Errorf("Expected NameError, got %T", err)
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
			t.Errorf("Expected LoadError, got %T", err)
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
			t.Errorf("Expected DuplicateNumberError, got %T", err)
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
		if len(history["test"].migrations) != 1 {
			t.Fatalf("expected one migration to be loaded")
		}
		if !history["test"].migrations[0].processed {
			t.Fatalf("expected migration state to be processed")
		}
	})
}

func TestLoadAppliedMigrations(t *testing.T) {
	app := gomodels.NewApp("test", "")
	if err := gomodels.Register(app); err != nil {
		t.Fatal(err)
	}
	defer gomodels.ClearRegistry()
	appState := &AppState{
		app:        gomodels.Registry()["test"],
		migrations: []*Node{},
	}
	history["test"] = appState
	defer clearHistory()
	err := gomodels.Start(gomodels.DBSettings{
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
			t.Fatalf("expected missing node error")
		}
		if mockedEngine.Calls("PrepareMigrations") != 1 {
			t.Fatalf("expected engine PrepareMigrations to be called")
		}
		if mockedEngine.Calls("GetMigrations") != 1 {
			t.Fatalf("expected engine GetMigrations to be called")
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
			t.Fatalf("expected migration to be applied")
		}
		mockedEngine.Reset()
	})
}
