package admin

import (
	"fmt"
	"reflect"

	"github.com/ecletus/core/resource"
	"github.com/go-aorm/aorm"
	"github.com/moisespsena-go/bintb"
	"github.com/moisespsena-go/maps"
)

type ResourceExportToCsvOptions struct {
	Search      func(crud *resource.CRUD, do func() (result interface{}, err error)) (result interface{}, err error)
	MakeChan    func() interface{}
	Data        maps.Map
	Name, Label string
}

func (this *Resource) ExportToCSV(opts ...*ResourceExportToCsvOptions) {
	const defaultName = "to_csv"
	var opt *ResourceExportToCsvOptions
	for _, opt = range opts {
	}
	if opt == nil {
		opt = &ResourceExportToCsvOptions{}
	}
	if opt.MakeChan == nil {
		opt.MakeChan = func() interface{} {
			return this.NewChanPtr(50)
		}
	}
	if opt.Search == nil {
		opt.Search = func(crud *resource.CRUD, do func() (result interface{}, err error)) (result interface{}, err error) {
			return do()
		}
	}
	if opt.Label == "" && opt.Name == "" {
		opt.Label = "CSV"
	}

	if opt.Name == "" {
		opt.Name = defaultName
	}

	ActionIndexBulkExport(this, &Action{
		Name:              opt.Name,
		Label:             opt.Label,
		PassCurrentParams: true,
		Data:              opt.Data,
	}, func(arg *ActionArgument) (err error) {
		arg.Context.Type |= INDEX
		var (
			result    = opt.MakeChan()
			ctx       = arg.Context
			C         = reflect.ValueOf(result).Elem()
			errResult = &struct {
				Done bool
				Err  error
			}{}
		)

		if err = ctx.Searcher.ParseContext(&Finder{
			Unlimited: true,
			FindMany: func(s *Searcher) (interface{}, error) {
				crud := DefaulSearchCrudFactory(s)
				crud.Chan = result
				return opt.Search(crud, func() (result interface{}, err error) {
					return crud.FindManyLayoutOrDefault(s.Layout)
				})
			},
		}); err != nil {
			if aorm.IsParseIdError(err) {
				return nil
			}
			return
		}

		if ctx.HasError() {
			return
		}

		go func() {
			_, errResult.Err = ctx.Searcher.FindMany()
			errResult.Done = true
		}()

		if _, ok := arg.Context.Request.URL.Query()["!plain"]; ok {
			arg.Context.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		} else {
			var prefix string
			if opt.Name != defaultName {
				prefix = opt.Name + "--"
			}
			arg.Context.Writer.Header().Set("Content-Disposition",
				fmt.Sprintf(`attachment; filename="%s--%s%s.csv"`,
					arg.Context.Ts(this.PluralLabelKey(), this.PluralName),
					prefix,
					ctx.RequestTime().Format("20060102-150405")))
			arg.Context.Writer.Header().Set("Content-Type", "text/csv; charset=utf-8")
		}

		arg.Context.RenderFlags |= CtxRenderCSV | CtxRenderEncode
		csvEnc := &CSVTransformer{
			Callbacks: CSVStreamWriterCallbacks{
				BeforeHeaderWrite: func(w *bintb.CsvStreamWriter, header *[]string) (err error) {
					*header = append(*header, "__URL__")
					return
				},
				AfterRecordEncoded: func(w *bintb.CsvStreamWriter, rec bintb.Recorde, values *[]string) (err error) {
					if w.RecordsCount() == 0 {
						*values = append(*values, arg.Context.OriginalURL.String())
					}
					return
				},
			},
		}

		encoder := &Encoder{
			Layout:   arg.Context.Layout,
			Resource: this,
			Context:  arg.Context,
			Result: func(i int) (r interface{}, err error) {
				if errResult.Err != nil {
					return nil, errResult.Err
				}
				if value, ok := C.Recv(); ok {
					if value.IsValid() {
						return value.Interface(), nil
					}
				}
				return nil, nil
			},
		}

		return csvEnc.Encode(arg.Context.Writer, encoder)
	})
}
