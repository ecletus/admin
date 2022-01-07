package admin

import (
	"net/url"
	"strings"

	"github.com/ecletus/core/resource"
	"github.com/moisespsena/template/html/template"
	tag_scanner "github.com/unapu-go/tag-scanner"
)

type ResourceIdConfig struct {
	SelectedTemplate   template.HTML
	SelectedTemplateJS *JS
	DisplayField       string
	DataResource
}

func (this *ResourceIdConfig) loadResourceConfig() {
	if this.SelectedTemplate == "" && this.SelectedTemplateJS == nil && this.DisplayField == "" {
		tags := this.Resource.UITags

		this.SelectedTemplate = template.HTML(tags.GetString("SELECT_TMPL"))
		this.DisplayField = tags.GetString("DISPLAY_FIELD")
		if js := tags.GetString("SELECTED_JS"); js != "" {
			this.SelectedTemplateJS = &JS{template.JS(js), true}
		}

		tags = this.Resource.Tags.Tags

		if this.SelectedTemplate == "" {
			this.SelectedTemplate = template.HTML(tags.GetString("UI_SELECT_TMPL"))
		}
		if this.DisplayField == "" {
			this.DisplayField = tags.GetString("DISPLAY_FIELD")
		}
		if this.SelectedTemplateJS == nil {
			if js := tags.GetString("SELECTED_JS"); js != "" {
				this.SelectedTemplateJS = &JS{template.JS(js), true}
			}
		}
	}
}

func (this *ResourceIdConfig) configureMeta(meta *Meta) {
	if meta.Type == "" {
		meta.Type = "resource_id"
	}
	this.Resource = meta.Resource

	if !meta.Tags.Flag("RESOURCE_ID") {
		if tags := meta.Tags.GetTags("RESOURCE_ID"); tags != nil {
			if resID := tags.GetString("RES"); resID != "" {
				meta.BaseResource.Admin.OnResourcesAdded(func(e *ResourceEvent) error {
					this.DataResource.Resource = e.Resource
					this.loadResourceConfig()
					return nil
				}, tags.GetString("RES"))
			}
			this.Scope(tags.GetTags("SCOPES", tag_scanner.FlagForceTags, tag_scanner.FlagNotNil, tag_scanner.FlagPreserveKeys).Flags()...)
			for key, value := range tags.GetTags("FILTERS", tag_scanner.FlagForceTags, tag_scanner.FlagNotNil, tag_scanner.FlagPreserveKeys) {
				this.Filter(key, value)
			}
			for key, value := range tags.GetTags("DEPS_VAL", tag_scanner.FlagForceTags, tag_scanner.FlagNotNil, tag_scanner.FlagPreserveKeys) {
				this.Dependency(&DependencyValue{Param: key, Value: value})
			}
		}
	}

	this.SelectedTemplate = template.HTML(this.Meta.UITags.GetString("SELECTED_TMPL"))
	this.DisplayField = this.Meta.UITags.GetString("DISPLAY_FIELD")
	if js := this.Meta.UITags.GetString("SELECTED_JS"); js != "" {
		this.SelectedTemplateJS = &JS{template.JS(js), true}
	}

	tags := this.Resource.Tags.Tags

	if this.SelectedTemplate == "" {
		this.SelectedTemplate = template.HTML(tags.GetString("UI_SELECT_TMPL"))
	}
	if this.DisplayField == "" {
		this.DisplayField = tags.GetString("DISPLAY_FIELD")
	}
	if this.SelectedTemplateJS == nil {
		if js := tags.GetString("SELECTED_JS"); js != "" {
			this.SelectedTemplateJS = &JS{template.JS(js), true}
		}
	}

	if this.Resource != nil {
		this.loadResourceConfig()
	}
}

func (this *ResourceIdConfig) ConfigureQorMeta(metaor resource.Metaor) {
	if this.Meta != nil {
		return
	}
	this.Meta = metaor.(*Meta)
	this.configureMeta(this.Meta)
}

func (this *ResourceIdConfig) GetUrl(context *Context, record interface{}) string {
	urlS := this.DataResource.URL(context)

	var params = url.Values{}

	params.Add(":no_actions", "true")

	if len(params) > 0 {
		if strings.ContainsRune(urlS, '?') {
			urlS += "&"
		} else {
			urlS += "?"
		}
		urlS += params.Encode()
	}

	return urlS
}

func init() {
	cfg := func(meta *Meta) {
		if meta.Config == nil {
			cfg := &ResourceIdConfig{}
			meta.Config = cfg
			cfg.ConfigureQorMeta(meta)
		}
	}
	RegisterMetaConfigor("resource_id", cfg)
	RegisterMetaPreConfigor(func(meta *Meta) {
		if meta.Tags.Tags["RESOURCE_ID"] != "" {
			meta.Type = "resource_id"
		}
	})
}
