package modelsx

import "github.com/volatiletech/sqlboiler/v4/queries/qm"

// QueryBuilder is a helper type to represent an array of qm.QueryMod objects
type QueryBuilder []qm.QueryMod

// NewBuilder creates a new, empty QueryBuilder
func NewBuilder() QueryBuilder {
	return make(QueryBuilder, 0)
}

// Add adds a qm.QueryMod object into the QueryBuilder (array)
func (qb QueryBuilder) Add(queries ...qm.QueryMod) QueryBuilder {
	qb = append(qb, queries...)
	return qb
}

// If adds a qm.QueryMod object to the QueryBuilder if a passed bool is true
func (qb QueryBuilder) IfCb(value bool, mods func() []qm.QueryMod) QueryBuilder {
	if value {
		qb = append(qb, mods()...)
	}
	return qb
}

// If adds a qm.QueryMod object to the QueryBuilder if a passed bool is true
func (qb QueryBuilder) If(value bool, mods ...qm.QueryMod) QueryBuilder {
	if value {
		qb = append(qb, mods...)
	}
	return qb
}
