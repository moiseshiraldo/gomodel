package migrations

import (
	"github.com/moiseshiraldo/gomodels"
)

var Migration = gomodels.New(
	"Migration",
	gomodels.Fields{
		"app":     gomodels.CharField{MaxLength: 50},
		"number":  gomodels.IntegerField{},
		"name":    gomodels.CharField{MaxLength: 100},
		"applied": gomodels.DateTimeField{AutoNowAdd: true},
	},
	gomodels.Options{},
)
