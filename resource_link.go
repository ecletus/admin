package admin

import "github.com/aghape/core"

func (res *Resource) GetIndexLink(context *core.Context, args ...interface{}) string {
	return res.GetLink(nil, context, args...)
}

func (res *Resource) GetLink(record interface{}, context *core.Context, args ...interface{}) string {
	var parentKeys []string
	for _, arg := range args {
		switch t := arg.(type) {
		case string:
			if t != "" {
				parentKeys = append(parentKeys, t)
			}
		case []string:
			parentKeys = append(parentKeys, t...)
		}
	}
	if record == nil {
		return res.GetContextIndexURI(context, parentKeys...)
	}
	uri := res.GetRecordURI(record, parentKeys...)
	return context.GenURL(uri)
}
