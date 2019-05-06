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

func (m Manager) GetQuerySet() QuerySet {
	cols := make([]string, 0, len(m.Model.fields))
	for name := range m.Model.fields {
		cols = append(cols, name)
	}
	constructor := m.Model.meta.Constructor
	if constructor == nil {
		constructor = Values{}
	}
	return GenericQuerySet{m.Model, constructor, "default", cols}
}

func (m Manager) All() QuerySet {
	return m.GetQuerySet()
}
