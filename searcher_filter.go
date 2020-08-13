package admin

import (
	"fmt"
	"mime/multipart"
	"net/url"
	"regexp"

	"github.com/ecletus/core/resource"
)

// Filter filter with defined filtersByName, filter with columns value
func (this *Searcher) Filter(filter *Filter, values *resource.MetaValues) {
	if this.filters == nil {
		this.filters = map[*Filter]*FilterArgument{}
	}
	this.filters[filter] = &FilterArgument{
		Filter:   filter,
		Scheme:   this.Scheme,
		Value:    values,
		Resource: this.Resource,
	}
}

// Filter filter with defined filtersByName, filter with columns value
func (this *Searcher) DefaultFilters() {
	if this.filters == nil {
		this.filters = map[*Filter]*FilterArgument{}
	}
	this.Scheme.Filters.EachDefaults(func(f *Filter) {
		if _, ok := this.filters[f]; !ok {
			this.filters[f] = &FilterArgument{
				Filter:   f,
				Scheme:   this.Scheme,
				Resource: this.Resource,
				Value:    &resource.MetaValues{},
				GoValue:  "",
			}
		}
	})
}

func (this *Searcher) FilterRaw(data map[string]string) {
	params := url.Values{}
	for key, value := range data {
		params.Add("filtersByName["+key+"].Value", value)
	}

	this.FilterFromParams(params, nil)
}

func (this *Searcher) FilterFromParams(params url.Values, form *multipart.Form) {
	for key := range params {
		if matches := filterRegexp.FindStringSubmatch(key); len(matches) > 0 {
			var prefix = fmt.Sprintf("filtersByName[%v].", matches[1])
			if filter := this.Scheme.Filters.Get(matches[1]); filter != nil {
				if metaValues, err := resource.ConvertFormDataToMetaValues(this.Context.Context, params, form, []resource.Metaor{}, prefix); err == nil {
					this.Filter(filter, metaValues)
				}
			}
		}
	}
}

func (this *Searcher) FilterRawPairs(args ...string) {
	data := make(map[string]string)
	l := len(args)
	for i := 0; i < l; i += 2 {
		data[args[i]] = args[i+1]
	}
	this.FilterRaw(data)
}

var filterRegexp = regexp.MustCompile(`^filtersByName\[(.*?)\]`)
