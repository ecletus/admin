package admin

import (
	"github.com/pkg/errors"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
)

func ParseMetaValueToRecord(record interface{}, metaValue *resource.MetaValue, context *core.Context) (v interface{}, err error) {
	meta := metaValue.Meta.(*Meta)
	if err = meta.Set(record, metaValue, context); err != nil {
		return nil, errors.Wrapf(err, "parse meta %q value", meta.Name)
	}
	return meta.Value(context, record), nil
}

func ParseMetaValue(metaValue *resource.MetaValue, context *core.Context) (v interface{}, err error) {
	return ParseMetaValueToRecord(metaValue.Meta.(*Meta).BaseResource.NewStruct(), metaValue, context)
}

func MustParseMetaValue(metaValue *resource.MetaValue, context *core.Context) (v interface{}) {
	var err error
	if v, err = ParseMetaValue(metaValue, context); err != nil {
		panic(err)
	}
	return
}

