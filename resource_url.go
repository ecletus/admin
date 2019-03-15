package admin

import "strings"

type ResourceURL struct {
	Resource       *Resource
	Scopes         []string
	Filters        map[string]string
	DynamicFilters func(context *Context, filters map[string]string)
	Layout         string
	Display        string
	Query          map[string]interface{}
	Dependencies   []interface{}
	recorde        bool
	FormatURI      func(data *ResourceURL, context *Context, uri string) string
	Scheme         string
	Suffix         string
}

func (url *ResourceURL) Dependency(dep ...interface{}) *ResourceURL {
	url.Dependencies = append(url.Dependencies, dep...)
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

// ToURLString Convert to URL string
func (url *ResourceURL) URL(context *Context) string {
	var parents []string
	var query []string

	if len(url.Dependencies) > 0 {
		for _, dep := range url.Dependencies {
			switch dp := dep.(type) {
			case *DependencyParent:
				if len(parents) == 0 {
					parents = make([]string, url.Resource.PathLevel, url.Resource.PathLevel)
				}
				parents[dp.Meta.Resource.PathLevel] = "{" + dp.Meta.Name + "}"
			case *DependencyQuery:
				query = append(query, dp.Param+"={"+dp.Meta.Name+"}")
			}
		}
	}

	if len(parents) > 0 {
		parent := url.Resource
		for pathLevel := url.Resource.PathLevel - 1; pathLevel >= 0; pathLevel-- {
			parent = parent.ParentResource
			if parents[pathLevel] == "" {
				parents[pathLevel] = context.URLParam(parent.ParamIDName())
			}
		}
	}

	var uri string
	if url.recorde {
		uri = url.Resource.GetContextURI(context.Context, "{ID}", parents...)
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

	if len(query) > 0 {
		uri += "?" + strings.Join(query, "&")
	}

	return uri
}
