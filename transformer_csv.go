package admin

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/moisespsena-go/bintb"
)

// CSVTransformer json transformer
type CSVTransformer struct {
	BeforeRecordeWrite,
	AfterRecordWrite func(w *bintb.CsvStreamWriter, rec bintb.Recorde) (err error)
}

// CouldEncode check if encodable
func (CSVTransformer) CouldEncode(encoder Encoder) bool {
	return true
}

// Encode encode encoder to writer as JSON
func (this CSVTransformer) Encode(writer io.Writer, encoder Encoder) (err error) {
	var (
		context    = encoder.Context
		res        = encoder.Resource
		metas      = context.convertSectionToMetas(res, context.indexSections(res))
		columns    = make([]*bintb.Column, len(metas), len(metas))
		formatters = make([]func(r interface{}) interface{}, len(metas), len(metas))
		csvW       = csv.NewWriter(writer)
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
			c.Tool, f = fmtr.(func(ctx *Context, m *Meta) (bintb.ColumnTool, func(r interface{}) interface{}))(context, m)
			formatters[i] = f
		}

		if c.Name == "" {
			c.Name = m.GetLabelC(context.Context)
		}
		columns[i] = c
	}

	if recv, ok := encoder.Result.(func(i int) (interface{}, error)); ok {
		var i int
		return bintb.CsvStreamWriteF(bintb.CsvStreamWriterOptions{
			OnlyColumnNames:    true,
			Encoder:            bintb.NewEncoder(columns),
			BeforeRecordeWrite: this.BeforeRecordeWrite,
			AfterRecordWrite:   this.AfterRecordWrite,
		}, csvW, func() (rec bintb.Recorde, err error) {
			var value interface{}
			if value, err = recv(i); err != nil || value == nil {
				return
			}
			i++
			rec = make(bintb.Recorde, len(metas))
			for i, m := range metas {
				if fmtr := formatters[i]; fmtr == nil {
					rec[i] = fmt.Sprint(context.FormattedValueOf(value, m))
				} else {
					rec[i] = fmtr(value)
				}
			}
			return
		})
	}
	return
}
