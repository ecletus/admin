package admin

import (
	"github.com/ecletus/core"
	"github.com/moisespsena-go/aorm"
)

func (this *Resource) GetIndexLink(context *core.Context, args ...interface{}) string {
	return this.GetLink(nil, context, args...)
}

func (this *Resource) GetLink(record interface{}, context *core.Context, args ...interface{}) string {
	var parentKeys = aorm.IDSlice(args...)
	if record == nil {
		return this.GetContextIndexURI(context, parentKeys...)
	}
	uri := this.GetRecordURI(record, parentKeys...)
	return context.Path(uri)
}
