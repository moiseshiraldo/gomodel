package migrations

import (
	"github.com/moiseshiraldo/gomodels"
)

func getModelChanges(model *gomodels.Model) OperationList {
	operations := OperationList{}
	app := model.App().Name()
	state := history[app]
	modelState, ok := state.models[model.Name()]
	if !ok {
		operation := &CreateModel{
			Name:   model.Name(),
			Fields: model.Fields(),
		}
		if model.Table() != model.App().Name()+model.Name() {
			operation.Table = model.Table()
		}
		operations = append(operations, operation)
		for idxName, fields := range model.Indexes() {
			operations = append(
				operations,
				&AddIndex{
					Model:  model.Name(),
					Name:   idxName,
					Fields: fields,
				},
			)
		}
	} else {
		for idxName := range modelState.Indexes() {
			if _, ok := model.Indexes()[idxName]; !ok {
				operations = append(
					operations,
					&RemoveIndex{Model: model.Name(), Name: idxName},
				)
			}
		}
		newFields := gomodels.Fields{}
		removedFields := []string{}
		for name := range modelState.Fields() {
			if _, ok := model.Fields()[name]; !ok {
				removedFields = append(removedFields, name)
			}
		}
		if len(removedFields) > 0 {
			operation := &RemoveFields{
				Model:  model.Name(),
				Fields: removedFields,
			}
			operations = append(operations, operation)
		}
		for name, field := range model.Fields() {
			if _, ok := modelState.Fields()[name]; !ok {
				newFields[name] = field
			}
		}
		if len(newFields) > 0 {
			operation := &AddFields{Model: model.Name(), Fields: newFields}
			operations = append(operations, operation)
		}
		for idxName, fields := range model.Indexes() {
			if _, ok := modelState.Indexes()[idxName]; !ok {
				operations = append(
					operations,
					&AddIndex{
						Model:  model.Name(),
						Name:   idxName,
						Fields: fields,
					},
				)
			}
		}
	}
	return operations
}
