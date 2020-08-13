package admin

import (
	"fmt"
	"strings"

	"github.com/ecletus/core/utils"
	"github.com/moisespsena-go/aorm"
	"github.com/moisespsena-go/maps"
)

type ResourceURL struct {
	Resource       *Resource
	Scopes         []string
	Filters        map[string]string
	DynamicFilters func(context *Context, filters map[string]string)
	Layout         string
	Display        string
	Query          maps.MapSI
	Dependencies   []interface{}
	recorde        bool
	FormatURI      func(data *ResourceURL, context *Context, uri string) string
	Scheme         string
	Suffix         string
}

func (url *ResourceURL) Dependency(dep ...interface{}) *ResourceURL {
	depLoop:
	for _, dep := range dep {
		switch dp := dep.(type) {
		case *DependencyParent:
			for i, other := range url.Dependencies {
				if parent, ok := other.(*DependencyParent); ok && parent.Meta.Name == dp.Meta.Name {
					url.Dependencies[i] = dp
					continue depLoop
				}
			}
		case *DependencyQuery:
			for i, other := range url.Dependencies {
				if parent, ok := other.(*DependencyQuery); ok && parent.Meta.Name == dp.Meta.Name {
					url.Dependencies[i] = dp
					continue depLoop
				}
			}
		case *DependencyValue:
			for i, other := range url.Dependencies {
				if parent, ok := other.(*DependencyValue); ok && parent.Param == dp.Param {
					url.Dependencies[i] = dp
					continue depLoop
				}
			}
		}
		url.Dependencies = append(url.Dependencies, dep)
	}
	return url
}

func (url *ResourceURL) Filter(name string, value string) *ResourceURL {
	if url.Filters == nil {
		url.Filters = make(map[string]string)
	}
	url.Filters[name] = value
	return url
}

func (url *ResourceURL) With(f func(r *ResourceURL)) *ResourceURL {
	f(url)
	return url
}

func (url *ResourceURL) Basic() *ResourceURL {
	url.Layout = BASIC_LAYOUT_HTML_WITH_ICON
	return url
}

// GenURL Convert to URL string using dependencies
func (url *ResourceURL) GenURL(context *Context, dependencies []interface{}) string {
	var parents []aorm.ID
	var query []string

	if len(dependencies) > 0 {
		for _, dep := range dependencies {
			switch dp := dep.(type) {
			case *DependencyParent:
				if len(parents) == 0 {
					parents = make([]aorm.ID, url.Resource.PathLevel, url.Resource.PathLevel)
				}
				if dp.Value != nil {
					parents[dp.Meta.Resource.PathLevel] = dp.Value
				} else {
					parents[dp.Meta.Resource.PathLevel] = aorm.FakeID("{" + dp.Meta.Name + "}")
				}
			case *DependencyQuery:
				query = append(query, dp.Param+"={"+dp.Meta.Name+"}")
			case *DependencyValue:
				query = append(query, dp.Param+"="+fmt.Sprint(dp.Value))
			}
		}
	}

	if len(parents) > 0 {
		parent := url.Resource
		for pathLevel := url.Resource.PathLevel - 1; pathLevel >= 0; pathLevel-- {
			parent = parent.ParentResource
			if parents[pathLevel].IsZero() {
				parents[pathLevel] = aorm.FakeID(context.URLParam(parent.ParamIDName()))
			}
		}
	}

	var uri string
	if url.recorde {
		uri = url.Resource.GetContextURI(context.Context, aorm.FakeID("{ID}"), parents...)
	} else {
		uri = url.Resource.GetContextIndexURI(context.Context, parents...)
	}

	if url.Scheme != "" {
		s := url.Resource.GetSchemeByName(url.Scheme)
		uri += s.Path()
	}

	uri += url.Suffix

	if url.FormatURI != nil {
		uri = url.FormatURI(url, context, uri)
	}

	if url.Layout != "" {
		query = append(query, P_LAYOUT+"="+url.Layout)
	}

	if url.Display != "" {
		query = append(query, P_DISPLAY+"="+url.Display)
	}

	for _, scope := range url.Scopes {
		query = append(query, "scopes="+scope)
	}

	for fname, fvalue := range url.Filters {
		query = append(query, "filtersByName["+fname+"].Value="+fvalue)
	}

	if url.DynamicFilters != nil {
		dynamicFilters := make(map[string]string)
		url.DynamicFilters(context, dynamicFilters)

		for fname, fvalue := range dynamicFilters {
			query = append(query, "filtersByName["+fname+"].Value="+fvalue)
		}
	}

	for name, value := range url.Query {
		switch t := value.(type) {
		case func(ctx *Context) (string, bool):
			if value, ok := t(context); ok {
				query = append(query, name+"="+value)
			}
		default:
			query = append(query, name+"="+utils.ToString(value))
		}
	}

	if len(query) > 0 {
		uri += "?" + strings.Join(query, "&")
	}

	return uri
}

// URL Convert to URL string
func (url *ResourceURL) URL(context *Context,dependencies ...interface{}) string {
	return url.GenURL(context, append(url.Dependencies, dependencies...))
}