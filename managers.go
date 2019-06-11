package gomodels

type Manager struct {
	Model *Model
}

func (m Manager) Create(values Values) (*Instance, error) {
	db := Databases["default"]
	container := m.Model.Container()
	instance := &Instance{m.Model, container}
	query, vals := sqlCreateQuery(m.Model, values, db.Driver)
	if db.Driver == "postgres" {
		var pk int64
		err := db.conn.QueryRow(query, vals...).Scan(&pk)
		if err != nil {
			return instance, &DatabaseError{
				"default",
				ErrorTrace{App: m.Model.app, Model: m.Model, Err: err},
			}
		}
		instance.Set(m.Model.pk, pk)
	} else {
		result, err := db.conn.Exec(query, vals...)
		if err != nil {
			return instance, &DatabaseError{
				"default",
				ErrorTrace{App: m.Model.app, Model: m.Model, Err: err},
			}
		}
		pk, err := result.LastInsertId()
		if err != nil {
			return instance, &DatabaseError{
				"default",
				ErrorTrace{App: m.Model.app, Model: m.Model, Err: err},
			}
		}
		instance.Set(m.Model.pk, pk)
	}
	for name, field := range m.Model.fields {
		if val, ok := values[name]; ok {
			if err := instance.Set(name, val); err != nil {
				return nil, err
			}
		} else if val, hasDefault := field.DefaultVal(); hasDefault {
			if err := instance.Set(name, val); err != nil {
				return nil, err
			}
		}
	}
	return instance, nil
}

func (m Manager) GetQuerySet() QuerySet {
	cols := make([]string, 0, len(m.Model.fields))
	for name := range m.Model.fields {
		cols = append(cols, name)
	}
	return GenericQuerySet{
		model:     m.Model,
		container: m.Model.meta.Container,
		database:  "default",
		columns:   cols,
	}
}

func (m Manager) All() QuerySet {
	return m.GetQuerySet()
}

func (m Manager) Filter(c Conditioner) QuerySet {
	return m.GetQuerySet().Filter(c)
}

func (m Manager) Get(c Conditioner) (*Instance, error) {
	return m.GetQuerySet().Get(c)
}

func (m Manager) SetContainer(container Container) QuerySet {
	return m.GetQuerySet().SetContainer(container)
}
