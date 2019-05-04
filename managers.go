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
	instance := Instance{m.Model, Values{m.Model.pk: pk}}
	for name, field := range m.Model.fields {
		val, ok := values[name]
		if ok {
			instance.Values[name] = val
		} else if hasDefault, defaultVal := field.DefaultVal(); hasDefault {
			instance.Values[name] = defaultVal
		}
	}
	return &instance, nil
}
