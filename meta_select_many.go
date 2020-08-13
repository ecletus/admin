package admin

import (
	"errors"
	"html/template"
	"reflect"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/ecletus/core/utils"
	"github.com/moisespsena-go/assetfs"
)

// SelectManyConfig meta configuration used for select many
type SelectManyConfig struct {
	Collection               interface{} // []string, [][]string, func(interface{}, *qor.Context) [][]string, func(interface{}, *admin.Context) [][]string
	DefaultCreating          bool
	Placeholder              string
	SelectionTemplate        string
	SelectMode               string // select, select_async, bottom_sheet
	Select2ResultTemplate    template.JS
	Select2SelectionTemplate template.JS
	RemoteDataResource       *DataResource
	SelectOneConfig
}

// GetTemplate get template for selection template
func (selectManyConfig SelectManyConfig) GetTemplate(context *Context, metaType string) (assetfs.AssetInterface, error) {
	if metaType == "form" && selectManyConfig.SelectionTemplate != "" {
		return context.Asset(selectManyConfig.SelectionTemplate)
	}
	return nil, errors.New("not implemented")
}

// ConfigureQorMeta configure select many meta
func (selectManyConfig *SelectManyConfig) ConfigureQorMeta(metaor resource.Metaor) {
	if meta, ok := metaor.(*Meta); ok {
		selectManyConfig.SelectOneConfig.Collection = selectManyConfig.Collection
		selectManyConfig.SelectOneConfig.SelectMode = selectManyConfig.SelectMode
		selectManyConfig.SelectOneConfig.DefaultCreating = selectManyConfig.DefaultCreating
		selectManyConfig.SelectOneConfig.Placeholder = selectManyConfig.Placeholder
		selectManyConfig.SelectOneConfig.RemoteDataResource = selectManyConfig.RemoteDataResource

		selectManyConfig.SelectOneConfig.ConfigureQorMeta(meta)

		selectManyConfig.RemoteDataResource = selectManyConfig.SelectOneConfig.RemoteDataResource
		selectManyConfig.DefaultCreating = selectManyConfig.SelectOneConfig.DefaultCreating
		meta.Type = "select_many"

		// Set FormattedValuer
		if meta.FormattedValuer == nil {
			meta.SetFormattedValuer(func(record interface{}, context *core.Context) interface{} {
				reflectValues := reflect.Indirect(reflect.ValueOf(meta.GetValuer()(record, context)))
				var results []string
				if reflectValues.IsValid() {
					for i := 0; i < reflectValues.Len(); i++ {
						results = append(results, utils.Stringify(reflectValues.Index(i).Interface()))
					}
				}
				return results
			})
		}
	}
}

func (selectManyConfig *SelectManyConfig) CurrentValues(ctx *Context, record interface{}, meta *Meta) interface{} {
	rawValue := meta.Value(ctx.Context, record)
	switch t := rawValue.(type) {
	case SelectManyValuesGetter:
		return t.Values()
	case SelectManyValuesContextGetter:
		return t.Values(ctx)
	default:
		return rawValue
	}
}

type SelectManyValuesGetter interface {
	Values() interface{}
}

type SelectManyValuesContextGetter interface {
	Values(ctx *Context) interface{}
}
