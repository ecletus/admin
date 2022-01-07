package admin

import "github.com/moisespsena-go/aorm"

type DefaultMetaTagger interface {
	AdminDefaultMetaTags() Tags
}

type DefaultFieldMetaTagger interface {
	AdminDefaultMetaTags(field *aorm.StructField, tags MetaTags) Tags
}
