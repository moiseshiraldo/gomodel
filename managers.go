package gomodels

type Manager struct {
	Model *Model
}

func (m Manager) Create(values Values) (*Instance, error) {
	db := Databases["default"]
	container := m.Model.Container()
	instance := &Instance{container, m.Model.meta.conType, m.Model}
	query, vals := sqlCreateQuery(m.Model.Table(), values)
	result, err := db.Exec(query, vals...)
	if err != nil {
		return instance, &DatabaseError{
			"default", ErrorTrace{App: m.Model.app, Model: m.Model, Err: err},
		}
	}
	pk, err := result.LastInsertId()
	if err != nil {
		return instance, &DatabaseError{
			"default", ErrorTrace{App: m.Model.app, Model: m.Model, Err: err},
		}
	}
	instance.Set(m.Model.pk, pk)
	for name, field := range m.Model.fields {
		val, ok := values[name]
		if ok {
			instance.Set(name, val)
		} else if hasDefault, defaultVal := field.DefaultVal(); hasDefault {
			instance.Set(name, defaultVal)
		}
	}
	return instance, nil
}

func (m Manager) GetQuerySet() QuerySet {
	cols := make([]string, 0, len(m.Model.fields))
	for name := range m.Model.fields {
		cols = append(cols, name)
	}
	container := m.Model.meta.Container
	return GenericQuerySet{
		model:     m.Model,
		container: m.Model.meta.Container,
		conType:   m.Model.meta.conType,
		database:  "default",
		columns:   cols,
	}
}

func (m Manager) All() QuerySet {
	return m.GetQuerySet()
}

func (m Manager) Filter(f Filterer) QuerySet {
	return m.GetQuerySet().Filter(f)
}

func (m Manager) Get(f Filterer) (*Instance, error) {
	return m.GetQuerySet().Get(f)
}

func (m Manager) SetContainer(container Container) QuerySet {
	return m.GetQuerySet().SetContainer(container)
}
