package migration

import (
	"github.com/moiseshiraldo/gomodel"
)

// Migration holds the model definition to store applied nodes in the database.
var Migration = gomodel.New(
	"Migration",
	gomodel.Fields{
		"app":     gomodel.CharField{MaxLength: 50},
		"number":  gomodel.IntegerField{},
		"name":    gomodel.CharField{MaxLength: 100},
		"applied": gomodel.DateTimeField{AutoNowAdd: true},
	},
	gomodel.Options{},
)
