package admin

import (
	"strings"
)

func (this *Admin) ParseResourceUrl(url string) (res *Resource, id string) {
	url = strings.TrimPrefix(url, "admin:/")
	return ParseResourceUrl(this.ResourcesByParam, url)
}

func (this *Resource) ParseResourceUrl(url string) (res *Resource, id string) {
	return ParseResourceUrl(this.ResourcesByParam, url)
}

func ParseResourceUrl(resourcesParam map[string]*Resource, url string) (res *Resource, id string) {
	url = strings.Trim(strings.TrimPrefix(url, "admin:/"), "/")
resources:
	for param, res_ := range resourcesParam {
		if strings.HasPrefix(url, param) {
			res = res_
			url = strings.TrimPrefix(strings.TrimPrefix(url, param), "/")
			if url != "" {
				if !res_.IsSingleton() {
					parts := strings.SplitN(url, "/", 2)
					id = parts[0]
					if len(parts) > 1 {
						url = parts[1]
						resourcesParam = res_.ResourcesByParam
						goto resources
					}
				} else {
					panic("not implemented")
				}
			}
			return
		}
	}
	return
}
