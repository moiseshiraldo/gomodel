package migrations

import (
	"github.com/moiseshiraldo/gomodels"
)

func getModelChanges(model *gomodels.Model) OperationList {
	operations := OperationList{}
	state := history[model.App().Name()]
	modelState, ok := state.Models[model.Name()]
	if !ok {
		operation := CreateModel{Model: model.Name(), Fields: model.Fields()}
		operations = append(operations, operation)
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
		for name, field := range model.Fields() {
			if _, ok := modelState.Fields()[name]; !ok {
				newFields[name] = field
			}
		}
		if len(newFields) > 0 {
			operation := AddFields{Model: model.Name(), Fields: newFields}
			operations = append(operations, operation)
		}
	}
	return operations
}
