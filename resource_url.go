package admin

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/moisespsena-go/maps"

	"github.com/ecletus/core/utils"
	"github.com/moisespsena-go/aorm"
)

type ResourceURL struct {
	Resource       *Resource
	Scopes         []string
	Filters        map[string]string
	FilterF        map[string]func(ctx *Context) string
	DynamicFilters func(context *Context, filters map[string]string)
	Layout         string
	Display        string
	Query          maps.MapSI
	Dependencies   []interface{}
	recorde        bool
	FormatURI      func(data *ResourceURL, context *Context, uri string) string
	Scheme         string
	Suffix         string
	URLHandlers    []func(ctx *Context, uri string, query *[]string)
}

func (this *ResourceURL) Scope(scope ...string) {
	this.Scopes = append(this.Scopes, scope...)
}

func (this *ResourceURL) Handler(f func(ctx *Context, uri string, query *[]string)) *ResourceURL {
	this.URLHandlers = append(this.URLHandlers, f)
	return this
}

func (this *ResourceURL) Dependency(dep ...interface{}) *ResourceURL {
depLoop:
	for _, dep := range dep {
		switch dp := dep.(type) {
		case *DependencyParent:
			for i, other := range this.Dependencies {
				if parent, ok := other.(*DependencyParent); ok && parent.Meta.Name == dp.Meta.Name {
					this.Dependencies[i] = dp
					continue depLoop
				}
			}
		case *DependencyQuery:
			for i, other := range this.Dependencies {
				if parent, ok := other.(*DependencyQuery); ok && parent.Meta.Name == dp.Meta.Name {
					this.Dependencies[i] = dp
					continue depLoop
				}
			}
		case *DependencyValue:
			for i, other := range this.Dependencies {
				if parent, ok := other.(*DependencyValue); ok && parent.Param == dp.Param {
					this.Dependencies[i] = dp
					continue depLoop
				}
			}
		}
		this.Dependencies = append(this.Dependencies, dep)
	}
	return this
}

func (this *ResourceURL) Filter(name string, value string) *ResourceURL {
	if this.Filters == nil {
		this.Filters = make(map[string]string)
	}
	this.Filters[name] = value
	return this
}

func (this *ResourceURL) FilterFunc(name string, value func(ctx *Context) string) *ResourceURL {
	if this.FilterF == nil {
		this.FilterF = make(map[string]func(ctx *Context) string)
	}
	this.FilterF[name] = value
	return this
}

func (this *ResourceURL) With(f func(r *ResourceURL)) *ResourceURL {
	f(this)
	return this
}

func (this *ResourceURL) Basic() *ResourceURL {
	this.Layout = BASIC_LAYOUT_HTML_WITH_ICON
	return this
}

// GenURL Convert to URL string using dependencies
func (this *ResourceURL) GenURL(context *Context, dependencies []interface{}) string {
	var (
		parents []aorm.ID
		query   []string
		e       = url.QueryEscape
	)

	if len(dependencies) > 0 {
		for _, dep := range dependencies {
			switch dp := dep.(type) {
			case *DependencyParent:
				if len(parents) == 0 {
					parents = make([]aorm.ID, this.Resource.PathLevel, this.Resource.PathLevel)
				}
				if dp.Value != nil {
					parents[dp.Meta.Resource.PathLevel] = dp.Value
				} else {
					parents[dp.Meta.Resource.PathLevel] = aorm.FakeID("{" + dp.Meta.Name + "}")
				}
			case *DependencyQuery:
				query = append(query, dp.Param+"={"+dp.Meta.Name+"}")
			case *DependencyValue:
				query = append(query, dp.Param+"="+e(fmt.Sprint(dp.Value)))
			}
		}
	}

	if len(parents) > 0 {
		parent := this.Resource
		for pathLevel := this.Resource.PathLevel - 1; pathLevel >= 0; pathLevel-- {
			parent = parent.ParentResource
			if parents[pathLevel].IsZero() {
				parents[pathLevel] = aorm.FakeID(context.URLParam(parent.ParamIDName()))
			}
		}
	}

	var uri string
	if this.recorde {
		uri = this.Resource.GetContextURI(context, aorm.FakeID("{ID}"), parents...)
	} else {
		uri = this.Resource.GetContextIndexURI(context, parents...)
	}

	if this.Scheme != "" {
		s := this.Resource.GetSchemeByName(this.Scheme)
		uri += s.Path()
	}

	uri += this.Suffix

	if this.FormatURI != nil {
		uri = this.FormatURI(this, context, uri)
	}

	if this.Layout != "" {
		query = append(query, P_LAYOUT+"="+e(this.Layout))
	}

	if this.Display != "" {
		query = append(query, P_DISPLAY+"="+e(this.Display))
	}

	for _, scope := range this.Scopes {
		query = append(query, "scope[]="+e(scope))
	}

	for fname, fvalue := range this.Filters {
		query = append(query, "filter["+fname+"].Value="+e(fvalue))
	}

	if this.FilterF != nil {
		for fname, fvalue := range this.FilterF {
			query = append(query, "filter["+fname+"].Value="+e(fvalue(context)))
		}
	}

	if this.DynamicFilters != nil {
		dynamicFilters := make(map[string]string)
		this.DynamicFilters(context, dynamicFilters)

		for fname, fvalue := range dynamicFilters {
			query = append(query, "filter["+fname+"].Value="+e(fvalue))
		}
	}

	for name, value := range this.Query {
		switch t := value.(type) {
		case func(ctx *Context) (string, bool):
			if value, ok := t(context); ok {
				query = append(query, name+"="+e(value))
			}
		default:
			query = append(query, name+"="+e(utils.ToString(value)))
		}
	}

	for _, handler := range this.URLHandlers {
		handler(context, uri, &query)
	}

	if len(query) > 0 {
		uri += "?" + strings.Join(query, "&")
	}

	return uri
}

// URL Convert to URL string
func (this *ResourceURL) URL(context *Context, dependencies ...interface{}) string {
	return this.GenURL(context, append(this.Dependencies, dependencies...))
}
