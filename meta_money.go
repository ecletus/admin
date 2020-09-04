package admin

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/leekchan/accounting"
	"github.com/pkg/errors"

	"github.com/ecletus/core"
	"github.com/ecletus/core/utils"
	"github.com/moisespsena-go/aorm/types"

	"github.com/moisespsena/template/html/template"

	"github.com/ecletus/core/resource"
)

type MoneyConfig struct {
	Locale           string
	Unsigned         bool
	Precision        int
	Zero             string
	RecordLocaleFunc func(ctx *core.Context, record interface{}) string
}

func (this *MoneyConfig) RecordLocale(ctx *core.Context, record interface{}) string {
	if this.RecordLocaleFunc != nil {
		return this.RecordLocaleFunc(ctx, record)
	}
	if this.Locale == "" {
		return this.DefaultLocale()
	}
	return this.Locale
}

func (MoneyConfig) DefaultLocale() string {
	return "BRL"
}

func (this *MoneyConfig) NumberFormatter(locale string) func(value float64) string {
	var (
		lc = accounting.LocaleInfo[locale]
		ac = accounting.Accounting{
			Symbol:    lc.ComSymbol,
			Precision: 2,
			Thousand:  lc.ThouSep,
			Decimal:   lc.DecSep,
		}
	)
	if this.Precision > 0 {
		ac.Precision = this.Precision
	}

	ac.Format = "%v"
	ac.FormatNegative = "-%v"

	return func(value float64) string {
		return accounting.FormatNumber(value, ac.Precision, ac.Thousand, ac.Decimal)
	}
}

func (this *MoneyConfig) FormattedValue(meta *Meta, record interface{}, context *core.Context) interface{} {
	value := meta.Value(context, record)
	if value == nil {
		return ""
	}

	if this.Zero != "" && meta.IsZero(record, value) {
		if this.Zero == "blank" {
			return ""
		}
		return this.Zero
	}

	var (
		lc = accounting.LocaleInfo[this.RecordLocale(context, record)]
		ac = accounting.Accounting{
			Symbol:    lc.ComSymbol,
			Precision: 2,
			Thousand:  lc.ThouSep,
			Decimal:   lc.DecSep,
		}
	)
	if this.Precision > 0 {
		ac.Precision = this.Precision
	}

	if lc.Pre {
		ac.Format = "%s %v"
		ac.FormatNegative = "%s -%v"
	} else {
		ac.Format = "%v %s"
		ac.FormatNegative = "-%v %s"
	}

	return ac.FormatMoney(value)
}

func (this *MoneyConfig) HtmlAttributtes(context *core.Context, record interface{}) template.HTML {
	var (
		lc        = accounting.LocaleInfo[this.RecordLocale(context, record)]
		precision = this.Precision
		attrs     = []string{`data-thousands="` + lc.ThouSep + `"`, `data-decimal="` + lc.DecSep + `"`}
	)
	if precision == 0 {
		precision = 2
	}
	attrs = append(attrs, `data-precision="`+strconv.Itoa(precision)+`"`)
	if lc.Pre {
		attrs = append(attrs, `data-prefix="`+lc.ComSymbol+` "`)
	} else {
		attrs = append(attrs, `data-suffix=" `+lc.ComSymbol+`"`)
	}
	if !this.Unsigned {
		attrs = append(attrs, `data-allow-negative="true"`)
	}
	return template.HTML(strings.Join(attrs, " "))
}

// ConfigureQorMeta configure select one meta
func (this *MoneyConfig) ConfigureQorMeta(metaor resource.Metaor) {
	meta := metaor.(*Meta)
	meta.Type = "money"

	tags := meta.Tags.GetTags("TYPE_OPT")
	if zero, ok := tags.GetOk("ZERO"); ok {
		this.Zero = zero
	}

	if meta.FormattedValuer == nil {
		meta.SetFormattedValuer(func(value interface{}, context *core.Context) interface{} {
			return this.FormattedValue(meta, value, context)
		})
	}
	if meta.Setter == nil {
		meta.Meta.Setter = resource.SingleFieldSetter(meta.Meta, meta.FieldName, func(_ bool, field reflect.Value, metaValue *resource.MetaValue, context *core.Context, record interface{}) (err error) {
			var values []float64
			if values, err = this.Parse(context, record, metaValue.FirstStringValue()); err != nil {
				return err
			}
			utils.SetNonZero(field, values[0])
			return nil
		})
	}
}

func (this *MoneyConfig) Parse(context *core.Context, record interface{}, value ...string) (result []float64, err error) {
	loc := this.RecordLocale(context, record)
	result = make([]float64, len(value))
	var precision = this.Precision
	if precision == 0 {
		precision = 2
	}
	for i, v := range value {
		if v == "" {
			continue
		}
		negative := v[0] == '-'
		if negative {
			v = strings.TrimSpace(v[1:])
			if v == "" {
				continue
			}
		}

		if result[i], err = strconv.ParseFloat(accounting.UnformatNumber(v, precision, loc), 64); err != nil {
			err = errors.Wrapf(err, "parse %q", v)
			return
		}
		if result[i] > 0 && negative {
			result[i] *= -1
		}
	}
	return
}

func init() {
	cfg := func(meta *Meta) {
		if meta.Config == nil {
			cfg := &MoneyConfig{}
			meta.Config = cfg
			cfg.ConfigureQorMeta(meta)
		}
	}
	RegisterMetaConfigor("money", cfg)
	RegisterMetaTypeConfigor(reflect.TypeOf(types.Money(0)), cfg)
}
