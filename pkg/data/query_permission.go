package data

import (
	"strings"

	"github.com/fatih/structs"
	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mytype"
	"github.com/marksauter/markus-ninja-api/pkg/util"
)

type QueryPermission struct {
	Operation mytype.Operation
	Audience  mytype.Audience
	Fields    pgtype.TextArray
}

func GetPermissableFields(model interface{}) (PermissableFields, error) {
	fields := structs.Fields(model)
	permissableFields := make([]*PermissableField, 0, len(fields))

	for _, field := range fields {
		permissableField, err := NewPermissableField(field)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		permissableFields = append(permissableFields, permissableField)
	}

	return PermissableFields(permissableFields), nil
}

type PermissableFields []*PermissableField

func (fs PermissableFields) Filter(al mytype.AccessLevel) PermissableFields {
	permissableFields := make([]*PermissableField, 0, len(fs))
	for _, f := range fs {
		if f.Can(al) {
			permissableFields = append(permissableFields, f)
		}
	}
	return PermissableFields(permissableFields)
}

func (fs PermissableFields) Names() []string {
	names := make([]string, len(fs))
	for i, f := range fs {
		names[i] = f.Name
	}
	return names
}

func NewPermissableField(f *structs.Field) (*PermissableField, error) {
	permissableField := &PermissableField{
		Name: f.Tag("db"),
	}
	permit := f.Tag("permit")
	if permit != "" {
		permissions := strings.Split(permit, "/")
		accessLookup := make(map[mytype.AccessLevel]bool, len(permissions))
		for _, p := range permissions {
			al, err := mytype.ParseAccessLevel(p)
			if err != nil {
				mylog.Log.WithError(err).Error(util.Trace(""))
				return nil, err
			}
			accessLookup[al] = true
		}
		permissableField.accessLookup = accessLookup
	}

	return permissableField, nil
}

type PermissableField struct {
	Name         string
	accessLookup map[mytype.AccessLevel]bool
}

func (fp *PermissableField) Can(al mytype.AccessLevel) bool {
	return fp.accessLookup[al]
}
