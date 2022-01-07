package admin

import (
	"context"
	"encoding/csv"
	"io"
	"reflect"

	strip "github.com/grokify/html-strip-tags-go"
	"github.com/moisespsena-go/bintb"
)

var csvTransformerKey contextKey = "csv_transformer"

type CSVStreamWriterCallbacks = bintb.CsvStreamWriterCallbacks

var CSVTransformerType = reflect.TypeOf(CSVTransformer{})

// CSVTransformer json transformer
type CSVTransformer struct {
	Callbacks CSVStreamWriterCallbacks
	Record    interface{}
}

// CouldEncode check if encodable
func (CSVTransformer) CouldEncode(*Encoder) bool {
	return true
}

func (this CSVTransformer) IsType(t reflect.Type) bool {
	return CSVTransformerType == t
}

// Encode encode encoder to writer as JSON
func (this CSVTransformer) Encode(w io.Writer, encoder *Encoder) (err error) {
	var (
		ctx        = encoder.Context.CreateChild(encoder.Resource, nil)
		secs       = ctx.indexSections()
		metas      = ctx.convertSectionToMetas(secs)
		columns    = make([]*bintb.Column, len(metas), len(metas))
		formatters = make([]func(r interface{}) interface{}, len(metas), len(metas))
		csvW       = csv.NewWriter(w)
	)
	csvW.UseCRLF = true
	csvW.Comma = ';'

	for i, m := range metas {
		c := &bintb.Column{
			Name: m.EncodedName,
			Tool: bintb.ColumnTypeTool.Get("str"),
		}

		if fmtr := m.Data.GetInterface(":csv"); fmtr != nil {
			var f func(r interface{}) interface{}
			c.Tool, f = fmtr.(func(ctx *Context, m *Meta) (bintb.ColumnTool, func(r interface{}) interface{}))(ctx, m)
			formatters[i] = f
		}

		if c.Name == "" {
			c.Name = m.GetLabelC(ctx.Context)
		}
		columns[i] = c
	}

	if recv, ok := encoder.Result.(func(i int) (interface{}, error)); ok {
		var (
			i   int
			opt = bintb.CsvStreamWriterOptions{
				Context:                  context.WithValue(ctx, csvTransformerKey, &this),
				OnlyColumnNames:          true,
				Encoder:                  bintb.NewEncoder(columns),
				CsvStreamWriterCallbacks: this.Callbacks,
			}
		)

		return bintb.CsvStreamWriteF(opt, csvW, func() (rec bintb.Recorde, err error) {
			if this.Record, err = recv(i); err != nil || this.Record == nil {
				return
			}
			i++
			rec = make(bintb.Recorde, len(metas))
			for i, m := range metas {
				if fmtr := formatters[i]; fmtr == nil {
					func() {
						mctx := &MetaContext{
							Meta:     m,
							Context:  ctx,
							Record:   this.Record,
							ReadOnly: true,
						}
						defer func() {
							for _, cb := range mctx.deferRenderHandlers {
								cb()
							}
						}()
						m.CallPrepareContextHandlers(mctx, this.Record)
						if v := m.FormattedValue(ctx.Context, this.Record); v == nil {
							rec[i] = ""
						} else if v.Value == "" && v.SafeValue != "" {
							rec[i] = strip.StripTags(v.SafeValue)
						} else {
							rec[i] = v.Value
						}
					}()
				} else {
					rec[i] = fmtr(this.Record)
				}
			}
			return
		})
	}
	return
}

func GetCsvTransformer(ctx context.Context) *CSVTransformer {
	return ctx.Value(csvTransformerKey).(*CSVTransformer)
}
