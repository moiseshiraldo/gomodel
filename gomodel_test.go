package gomodel

import (
	"testing"
)

// TestModels tests the Model struct methods
func TestModel(t *testing.T) {
	app := &Application{name: "users", models: map[string]*Model{}}
	model := &Model{
		name: "User",
		app:  app,
		fields: Fields{
			"email": CharField{MaxLength: 100},
		},
		meta: Options{
			Indexes: Indexes{"test_idx": []string{"email"}},
		},
	}

	t.Run("Name", func(t *testing.T) {
		if model.Name() != "User" {
			t.Errorf("expected User, got %s", model.Name())
		}
	})

	t.Run("App", func(t *testing.T) {
		if model.App() != app {
			t.Error("method returned wrong app")
		}
	})

	t.Run("DefaultTable", func(t *testing.T) {
		model.meta.Table = ""
		if model.Table() != "users_user" {
			t.Errorf("expected users_user, got %s", model.Table())
		}
	})

	t.Run("CustomTable", func(t *testing.T) {
		model.meta.Table = "custom_table"
		if model.Table() != "custom_table" {
			t.Errorf("expected custom_table, got %s", model.Table())
		}
	})

	t.Run("Fields", func(t *testing.T) {
		fields := model.Fields()
		if _, ok := fields["email"]; !ok {
			t.Fatal("result is missing email field")
		}
		delete(fields, "email")
		if _, ok := model.fields["email"]; !ok {
			t.Error("internal model fields map was modified")
		}
	})

	t.Run("Indexes", func(t *testing.T) {
		indexes := model.Indexes()
		if _, ok := indexes["test_idx"]; !ok {
			t.Fatal("result is missing test_idx")
		}
		delete(indexes, "test_idx")
		if _, ok := model.meta.Indexes["test_idx"]; !ok {
			t.Error("internal model indexes map was modified")
		}
	})

	t.Run("BuilderContainer", func(t *testing.T) {
		model.meta.Container = Values{"email": "user@test.com"}
		container := model.Container()
		if _, ok := container.(Values); !ok {
			t.Fatalf("expected Values type, got %T", container)
		}
		if len(container.(Values)) > 0 {
			t.Error("expected container to be empty")
		}
	})

	t.Run("StructContainer", func(t *testing.T) {
		type userContainer struct {
			email string
		}
		model.meta.Container = &userContainer{email: "user@test.com"}
		container := model.Container()
		if _, ok := container.(*userContainer); !ok {
			t.Fatalf("expected userContainer type, got %T", container)
		}
		if container.(*userContainer).email != "" {
			t.Error("expected container to be empty")
		}
	})

	t.Run("DuplicatePK", func(t *testing.T) {
		model.pk = ""
		model.fields["id"] = IntegerField{PrimaryKey: true}
		model.fields["username"] = CharField{PrimaryKey: true}
		if err := model.SetupPrimaryKey(); err == nil {
			t.Error("expected duplicate primery key error")
		}
	})

	t.Run("FieldNamePK", func(t *testing.T) {
		model.pk = ""
		model.fields["id"] = IntegerField{PrimaryKey: true}
		model.fields["pk"] = CharField{}
		if err := model.SetupPrimaryKey(); err == nil {
			t.Error("expected reserved field name pk error")
		}
	})

	t.Run("SkipPKSetup", func(t *testing.T) {
		model.pk = "id"
		if err := model.SetupPrimaryKey(); err != nil {
			t.Error(err)
		}
	})

	t.Run("ManualPKSetup", func(t *testing.T) {
		model.pk = ""
		model.fields = Fields{"email": CharField{PrimaryKey: true}}
		if err := model.SetupPrimaryKey(); err != nil {
			t.Fatal(err)
		}
		if model.pk != "email" {
			t.Errorf("expected pk to be email, got %s", model.pk)
		}
	})

	t.Run("AutoPKSetup", func(t *testing.T) {
		model.pk = ""
		model.fields = Fields{}
		if err := model.SetupPrimaryKey(); err != nil {
			t.Fatal(err)
		}
		if model.pk != "id" {
			t.Fatalf("expected pk to be id, got %s", model.pk)
		}
		if _, ok := model.fields["id"]; !ok {
			t.Fatal("expected model to contain id field")
		}
		field := model.fields["id"]
		if _, ok := field.(IntegerField); !ok {
			t.Error("expected id field to be IntegerField")
		}
		if !field.IsPK() {
			t.Error("expeceted id field to be primary key")
		}
	})

	t.Run("EmptyIndex", func(t *testing.T) {
		model.fields = Fields{}
		model.meta.Indexes = Indexes{
			"empty_idx": []string{},
		}
		if err := model.SetupIndexes(); err == nil {
			t.Error("expected empty index error")
		}
	})

	t.Run("DuplicateIndex", func(t *testing.T) {
		model.fields = Fields{"email": CharField{Index: true}}
		model.meta.Indexes = Indexes{
			"users_user_email_auto_idx": []string{"email"},
		}
		if err := model.SetupIndexes(); err == nil {
			t.Error("expected duplicate index error")
		}
	})

	t.Run("IndexUnknownField", func(t *testing.T) {
		model.fields = Fields{}
		model.meta.Indexes = Indexes{
			"invalid_indx": []string{"username"},
		}
		if err := model.SetupIndexes(); err == nil {
			t.Error("expected duplicate index error")
		}
	})

	t.Run("SetupIndexes", func(t *testing.T) {
		model.fields = Fields{"username": CharField{Index: true}}
		model.meta.Indexes = Indexes{}
		if err := model.SetupIndexes(); err != nil {
			t.Fatal(err)
		}
		if _, ok := model.meta.Indexes["users_user_username_auto_idx"]; !ok {
			t.Fatal("index was not added to model")
		}
		fields := model.meta.Indexes["users_user_username_auto_idx"]
		if len(fields) == 0 || fields[0] != "username" {
			t.Error("added index with wrong details")
		}
	})

	t.Run("RegisterDuplicate", func(t *testing.T) {
		app.models["User"] = model
		if err := model.Register(app); err == nil {
			t.Error("expected duplicate model error")
		}
	})

	t.Run("RegisterInvalidPK", func(t *testing.T) {
		app.models = map[string]*Model{}
		model.fields = Fields{
			"id":    IntegerField{PrimaryKey: true},
			"email": CharField{PrimaryKey: true},
		}
		model.pk = ""
		if err := model.Register(app); err == nil {
			t.Error("expected duplicate pk error")
		}
	})

	t.Run("RegisterInvalidIndexes", func(t *testing.T) {
		app.models = map[string]*Model{}
		model.fields = Fields{}
		model.meta.Indexes = Indexes{"username": []string{}}
		if err := model.Register(app); err == nil {
			t.Error("expected empty index error")
		}
	})

	t.Run("AddDuplicateField", func(t *testing.T) {
		model.fields = Fields{"email": CharField{}}
		if err := model.AddField("email", CharField{}); err == nil {
			t.Error("expected duplicate field error")
		}
	})

	t.Run("AddFieldDuplicatePK", func(t *testing.T) {
		model.pk = "id"
		model.fields = Fields{}
		field := CharField{PrimaryKey: true}
		if err := model.AddField("foo", field); err == nil {
			t.Error("expected duplicate pk error")
		}
	})

	t.Run("AddField", func(t *testing.T) {
		model.fields = Fields{}
		if err := model.AddField("foo", CharField{}); err != nil {
			t.Fatal(err)
		}
		if _, ok := model.fields["foo"]; !ok {
			t.Error("field was not added to model")
		}
	})

	t.Run("RemovePKField", func(t *testing.T) {
		model.pk = "id"
		model.fields = Fields{"id": IntegerField{PrimaryKey: true}}
		if err := model.RemoveField("id"); err == nil {
			t.Error("expected cannot remove pk error")
		}
	})

	t.Run("RemoveUnknownField", func(t *testing.T) {
		model.fields = Fields{}
		if err := model.RemoveField("bar"); err == nil {
			t.Error("expected field does not exist error")
		}
	})

	t.Run("RemoveIndexesField", func(t *testing.T) {
		model.fields = Fields{"email": CharField{}}
		model.meta.Indexes = Indexes{"email_idx": []string{"email"}}
		if err := model.RemoveField("email"); err == nil {
			t.Error("expected cannot remove indexed field error")
		}
	})

	t.Run("RemoveField", func(t *testing.T) {
		model.fields = Fields{"email": CharField{}}
		model.meta.Indexes = Indexes{}
		if err := model.RemoveField("email"); err != nil {
			t.Fatal(err)
		}
		if _, found := model.fields["email"]; found {
			t.Error("field was not removed from model")
		}
	})

	t.Run("AddDuplicateIndex", func(t *testing.T) {
		model.meta.Indexes = Indexes{"email_idx": []string{"email"}}
		if err := model.AddIndex("email_idx", "email"); err == nil {
			t.Error("expected duplicate index error")
		}
	})

	t.Run("AddEmptyIndex", func(t *testing.T) {
		model.meta.Indexes = Indexes{}
		if err := model.AddIndex("email_idx"); err == nil {
			t.Error("expected empty index error")
		}
	})

	t.Run("AddIndexUnknownField", func(t *testing.T) {
		model.meta.Indexes = Indexes{}
		model.fields = Fields{}
		if err := model.AddIndex("test_idx", "foo", "bar"); err == nil {
			t.Error("expected field not found error")
		}
	})

	t.Run("AddIndex", func(t *testing.T) {
		model.meta.Indexes = Indexes{}
		model.fields = Fields{"email": CharField{}}
		if err := model.AddIndex("test_idx", "email"); err != nil {
			t.Fatal(err)
		}
		if _, ok := model.meta.Indexes["test_idx"]; !ok {
			t.Fatal("index was not added to model")
		}
		fields := model.meta.Indexes["test_idx"]
		if len(fields) == 0 || fields[0] != "email" {
			t.Error("index added with wrong details")
		}
	})

	t.Run("RemoveUnknownIndex", func(t *testing.T) {
		model.meta.Indexes = Indexes{}
		if err := model.RemoveIndex("foo"); err == nil {
			t.Error("expected index not found error")
		}
	})

	t.Run("RemoveUnknownIndex", func(t *testing.T) {
		model.meta.Indexes = Indexes{"foo": []string{"bar"}}
		if err := model.RemoveIndex("foo"); err != nil {
			t.Fatal(err)
		}
		if _, found := model.meta.Indexes["foo"]; found {
			t.Error("index was not removed from model")
		}
	})
}

// TestDispatcher test the Dispatcher struct methods
func TestDispatcher(t *testing.T) {
	model := &Model{
		name: "User",
		fields: Fields{
			"foo": CharField{Default: "bar"},
		},
		meta: Options{Container: Values{}},
	}
	dispatcher := Dispatcher{
		Model:   model,
		Objects: Manager{Model: model, QuerySet: GenericQuerySet{}},
	}

	t.Run("NewInstance", func(t *testing.T) {
		instance, err := dispatcher.New(Values{"foo": "test"})
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := instance.container.(Values); !ok {
			t.Fatalf(
				"expected instance container to be Values, got %T",
				instance.container,
			)
		}
		if val := instance.container.(Values)["foo"]; val != "test" {
			t.Fatalf("expected field value to be bar, got %s", val)
		}
	})

	t.Run("NewInstanceInvalidField", func(t *testing.T) {
		type userContainer struct {
			email string
		}
		model.meta.Container = userContainer{}
		_, err := dispatcher.New(Values{"foo": "test"})
		if _, ok := err.(*ContainerError); !ok {
			t.Errorf("expected ContainerError, got %T", err)
		}

	})

	t.Run("NewInstanceInvalidDefault", func(t *testing.T) {
		type userContainer struct {
			email string
		}
		model.meta.Container = userContainer{}
		_, err := dispatcher.New(Values{})
		if _, ok := err.(*ContainerError); !ok {
			t.Errorf("expected ContainerError, got %T", err)
		}

	})

	t.Run("NewModel", func(t *testing.T) {
		dispatcher := New(
			"User",
			Fields{
				"username": CharField{MaxLength: 150},
			},
			Options{},
		)
		if dispatcher.Model.name != "User" {
			t.Errorf(
				"expected model name to be User, got %s", dispatcher.Model.name,
			)
		}
		if _, ok := dispatcher.Model.fields["username"]; !ok {
			t.Error("model is missing username field")
		}
		manager := dispatcher.Objects
		if manager.Model != dispatcher.Model {
			t.Error("manager was not set up correctly")
		}
	})
}
