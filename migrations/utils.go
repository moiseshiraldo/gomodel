package migrations

import (
	"database/sql"
	"github.com/moiseshiraldo/gomodels"
)

func getModelChanges(model *gomodels.Model) OperationList {
	operations := OperationList{}
	app := model.App().Name()
	state := history[app]
	modelState, ok := state.Models[model.Name()]
	if !ok {
		operations = append(
			operations,
			CreateModel{
				Name:   model.Name(),
				Fields: model.Fields(),
			},
		)
		for idxName, columns := range model.Indexes() {
			operations = append(
				operations,
				AddIndex{
					Model:   model.Name(),
					Name:    idxName,
					Columns: columns,
				},
			)
		}
	} else {
		newFields := gomodels.Fields{}
		removedFields := []string{}
		for name := range modelState.Fields() {
			if _, ok := model.Fields()[name]; !ok {
				removedFields = append(removedFields, name)
			}
		}
		if len(removedFields) > 0 {
			operation := RemoveFields{
				Model:  model.Name(),
				Fields: removedFields,
			}
			operations = append(operations, operation)
		}
		for idxName := range modelState.Indexes() {
			if _, ok := model.Indexes()[idxName]; !ok {
				operations = append(
					operations,
					RemoveIndex{Model: model.Name(), Name: idxName},
				)
			}
		}
		for name, field := range model.Fields() {
			if _, ok := modelState.Fields()[name]; !ok {
				newFields[name] = field
			}
		}
		if len(newFields) > 0 {
			operation := AddFields{Model: model.Name(), Fields: newFields}
			operations = append(operations, operation)
		}
		for idxName, columns := range model.Indexes() {
			if _, ok := modelState.Indexes()[idxName]; !ok {
				operations = append(
					operations,
					AddIndex{
						Model:   model.Name(),
						Name:    idxName,
						Columns: columns,
					},
				)
			}
		}
	}
	return operations
}

func prepareDatabase(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS gomodels_Migration (
		'id' integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		'app' varchar(50) NOT NULL,
		'name' varchar(100) NOT NULL,
		'number' integer NOT NULL
	)`
	if _, err := db.Exec(query); err != nil {
		return err
	}
	return nil
}
