package admin

import (
	"github.com/moisespsena-go/aorm"
)

func (this *Resource) GetIndexLink(context *Context, args ...interface{}) string {
	return this.GetLink(context, nil, args...)
}

func (this *Resource) GetLink(context *Context, record interface{}, args ...interface{}) string {
	var parentKeys = aorm.IDSlice(args...)
	if record == nil {
		return this.GetContextIndexURI(context, parentKeys...)
	}
	if uri := this.GetRecordURI(context, record, parentKeys...); uri != "" {
		return context.Path(uri)
	}
	return ""
}
