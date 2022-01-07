package admin

import "github.com/moisespsena-go/getters"

type (
	Stringer interface {
		AdminString(ctx *Context, opt getters.Getter) string
	}

	Equaler interface {
		Equals(value interface{}) bool
	}

	SoftDeleter interface {
		CanSoftDeleter() bool
		IsSoftDeleted() bool
	}
)
