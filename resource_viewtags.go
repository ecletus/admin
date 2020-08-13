package admin

func (this *Scheme) SetViewTagger(tagger ResourceViewTager) {
	this.SchemeData.Set("view_tagger", tagger)
}

func (this *Resource) GetViewTags(ctx *Context, record interface{}) []string {
	if ctx.Scheme != nil {
		if data, ok := ctx.Scheme.SchemeData.Get("view_tagger"); ok {
			return data.(ResourceViewTager).Tags(ctx, record)
		}
	}
	if data, ok := this.Scheme.SchemeData.Get("view_tagger"); ok {
		return data.(ResourceViewTager).Tags(ctx, record)
	}
	return []string{}
}

type ResourceViewTager interface {
	Tags(ctx *Context, record interface{}) []string
}

type ResourceViewTagsFunc = func(ctx *Context, record interface{}) []string
type funcResourceViewTags ResourceViewTagsFunc

func (this funcResourceViewTags) Tags(ctx *Context, record interface{}) []string {
	return this(ctx, record)
}

func NewResourceViewTags(f ResourceViewTagsFunc) ResourceViewTager {
	return funcResourceViewTags(f)
}
