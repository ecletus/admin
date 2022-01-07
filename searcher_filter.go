package admin

import (
	"fmt"
	"mime/multipart"
	"net/url"
	"regexp"
	"strings"

	"github.com/ecletus/core/resource"
)

// Filter filter with defined filter, filter with columns value
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

func (this *Searcher) GetFilters(advanced bool) (res []*FilterArgument) {
	for _, f := range this.Filters {
		if f.Filter.IsAdvanced() == advanced {
			res = append(res, f)
		}
	}
	return
}

func (this *Searcher) HasFilters(advanced bool) bool {
	for _, f := range this.Filters {
		if f.Filter.IsAdvanced() == advanced {
			return true
		}
	}
	return false
}

func (this *Searcher) CountFilters(advanced bool) (i int) {
	for _, f := range this.Filters {
		if f.Filter.IsAdvanced() == advanced {
			i++
		}
	}
	return
}

// Filter filter with defined filter, filter with columns value
func (this *Searcher) DefaultFilters() {
	if this.filters == nil {
		this.filters = map[*Filter]*FilterArgument{}
	}
	this.Scheme.Filters.EachDefaults(func(f *Filter) {
		if _, ok := this.filters[f]; !ok {
			this.filters[f] = &FilterArgument{
				Default:  true,
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
		params.Add("filter["+key+"].Value", value)
	}

	this.FilterFromParams(params, nil)
}

func (this *Searcher) FilterFromParams(params url.Values, form *multipart.Form) {
	for key := range params {
		if matches := filterRegexp.FindStringSubmatch(key); len(matches) > 0 {
			var prefix = fmt.Sprintf("filter[%v].", matches[1])
			if filter := this.Scheme.Filters.GetByName(matches[1]); filter != nil {
				if metaValues, err := resource.ConvertFormDataToMetaValues(this.Context.Context, params, form, []resource.Metaor{}, prefix); err == nil {
					this.Filter(filter, metaValues)
				}
			}
		} else if matches = advFilterRegexp.FindStringSubmatch(key); len(matches) > 0 {
			var prefix = fmt.Sprintf("adv_filter[%v].", matches[1])
			if filter := this.Scheme.Filters.GetByName(matches[1]); filter != nil {
				if metaValues, err := resource.ConvertFormDataToMetaValues(this.Context.Context, params, form, []resource.Metaor{}, prefix); err == nil {
					if !metaValues.IsBlank() {
						this.Filter(filter, metaValues)
					}
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

var (
	filterRegexp    = regexp.MustCompile(`^filter\[(.*?)\]`)
	advFilterRegexp = regexp.MustCompile(`^adv_filter\[(.*?)\]`)
)

func GetFilterFromQS(queryKey string) (name, key string) {
	m := filterRegexp.FindAllStringSubmatch(queryKey, 1)
	if m != nil {
		return m[0][1], strings.TrimPrefix(queryKey, m[0][0])[1:]
	}
	return
}

func GetAdvFilterFromQS(queryKey string) (name, key string) {
	m := advFilterRegexp.FindAllStringSubmatch(queryKey, 1)
	if m != nil {
		return m[0][1], strings.TrimPrefix(queryKey, m[0][0])[1:]
	}
	return
}
