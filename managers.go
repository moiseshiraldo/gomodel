package gomodels

type Manager struct {
	Model *Model
}

func (m Manager) Create(values Values) (Constructor, error) {
	db := Databases["default"]
	query, vals := sqlCreateQuery(m.Model.Table(), values)
	result, err := db.Exec(query, vals...)
	if err != nil {
		return nil, &DatabaseError{
			"default", ErrorTrace{App: m.Model.app, Model: m.Model, Err: err},
		}
	}
	pk, err := result.LastInsertId()
	if err != nil {
		return nil, &DatabaseError{
			"default", ErrorTrace{App: m.Model.app, Model: m.Model, Err: err},
		}
	}
	instance := Instance{Values{m.Model.pk: pk}, m.Model}
	for name, field := range m.Model.fields {
		val, ok := values[name]
		if ok {
			instance.Set(name, val)
		} else if hasDefault, defaultVal := field.DefaultVal(); hasDefault {
			instance.Set(name, defaultVal)
		}
	}
	return &instance, nil
}
