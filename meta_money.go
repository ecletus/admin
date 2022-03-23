package admin

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/ecletus/core/helpers"
	"github.com/ecletus/validations"
	"github.com/leekchan/accounting"
	"github.com/moisespsena-go/aorm"
	"github.com/moisespsena-go/aorm/types/money"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"golang.org/x/text/message"

	"github.com/ecletus/core"
	"github.com/ecletus/core/utils"
	"github.com/moisespsena/template/html/template"

	"github.com/ecletus/core/resource"
)

type MoneyConfig struct {
	Locale           string
	Sig              *money.Signal
	Precision        int
	Zero             string
	RecordLocaleFunc func(ctx *core.Context, record interface{}) string
	meta             *Meta
	NewValue         resource.GetFielder
}

func (this *MoneyConfig) RecordLocale(ctx *core.Context, record interface{}) string {
	if record != nil && this.RecordLocaleFunc != nil {
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

func (this *MoneyConfig) NumberFormatter(locale string) func(value money.Money) string {
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

	return func(value money.Money) string {
		return accounting.FormatNumber(value.Decimal, ac.Precision, ac.Thousand, ac.Decimal)
	}
}

func (this *MoneyConfig) ToStringInterface(context *core.Context, record, value interface{}) string {
	return MoneyToString(this.RecordLocale(context, record), this.Precision, value)
}

func (this *MoneyConfig) FormattedValue(meta *Meta, record interface{}, context *core.Context) *FormattedValue {
	value := meta.Value(context, record)
	if value == nil {
		return nil
	}

	if value.(helpers.Zeroer).IsZero() {
		if this.Zero != "" && this.Zero != "blank" {
			return &FormattedValue{Record: record, Raw: money.Zero, Value: this.Zero}
		}
		if !meta.Tags.ZeroRender() {
			return nil
		}
	}

	fv := &FormattedValue{Record: record, Raw: value}
	ctx := ContextFromCoreContext(context)

	if enc := ctx.Encoder(); enc != nil && enc.IsType(CSVTransformerType) {
		fv.Value = message.NewPrinter(*ctx.LangTag).Sprint(value)
	} else {
		fv.Value = this.ToStringInterface(context, record, value)
	}
	return fv
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
	if this.Sig != nil {
		sign := *this.Sig
		if sign == money.ALL || sign == money.NEGATIVE {
			attrs = append(attrs, `data-allow-negative="true"`)
		}
	} else {
		attrs = append(attrs, `data-allow-negative="true"`)
	}
	return template.HTML(strings.Join(attrs, " "))
}

// ConfigureQorMeta configure select one meta
func (this *MoneyConfig) ConfigureQorMeta(metaor resource.Metaor) {
	meta := metaor.(*Meta)
	this.meta = meta
	meta.Type = "money"

	tags := meta.Tags.GetTags("TYPE_OPT")
	if zero, ok := tags.GetOk("ZERO"); ok {
		this.Zero = zero
	}

	if this.Sig == nil && meta.FieldStruct != nil {
		sign := money.GetFieldConfig(meta.FieldStruct).Sig
		if sign != money.ALL {
			this.Sig = &sign
		}
	}

	if meta.FormattedValuer == nil {
		meta.SetFormattedValuer(func(value interface{}, context *core.Context) *FormattedValue {
			return this.FormattedValue(meta, value, context)
		})
	}
	if meta.Setter == nil || this.NewValue != nil {
		setter := func(_ bool, field reflect.Value, metaValue *resource.MetaValue, context *core.Context, record interface{}) (err error) {
			var values []money.Money
			if values, err = this.Parse(context, record, metaValue.StringValue()); err != nil {
				return err
			}
			v := values[0]
			if !v.IsZero() && this.Sig != nil {
				if sign := *this.Sig; v.Sign() != int(sign) {
					switch sign {
					case money.NEGATIVE:
						return meta.MakeError(context, record, validations.NewStringError("Apenas valor NEGATIVO"))
					case money.POSITIVE:
						return meta.MakeError(context, record, validations.NewStringError("Apenas valor POSITIVO"))
					}
				}
			}
			utils.SetNonZero(field, values[0])
			return nil
		}
		if this.NewValue == nil {
			meta.Meta.Setter = resource.SingleFieldSetter(meta.FieldName, setter)
		} else {
			meta.Meta.Setter = resource.Setter(this.NewValue, setter)
		}
	}
}

func (this *MoneyConfig) Parse(context *core.Context, record interface{}, value ...string) (result []money.Money, err error) {
	loc := this.RecordLocale(context, record)
	result = make([]money.Money, len(value))
	var (
		precision = this.Precision
		d         decimal.Decimal
	)
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

		if d, err = decimal.NewFromString(accounting.UnformatNumber(v, precision, loc)); err != nil {
			err = errors.Wrapf(err, "parse %q", v)
			return
		}

		if d.IsPositive() && negative {
			d = d.Neg()
		}

		result[i] = money.Money{Decimal: d}
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
	RegisterMetaTypeConfigor(reflect.TypeOf(money.Money{}), cfg)
}

func MoneyToString(locale string, precision int, value interface{}) string {
	if precision == 0 {
		precision = 2
	}
	var (
		lc = accounting.LocaleInfo[locale]
		ac = accounting.Accounting{
			Symbol:    lc.ComSymbol,
			Precision: precision,
			Thousand:  lc.ThouSep,
			Decimal:   lc.DecSep,
		}
	)

	switch t := value.(type) {
	case decimal.Decimal:
	case *decimal.Decimal:
		value = *t
	case aorm.ToDecimalConverter:
		value = t.ToDecimal()
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

func MoneyTagsOf(v money.Money) string {
	if v.IsZero() {
		return "money-zero"
	}
	if v.IsNegative() {
		return "money-debit"
	}
	return "money-credit"
}

func MoneyTagsOfDst(v money.Money, dst *[]string) {
	tag := MoneyTagsOf(v)
	if tag != "" {
		*dst = append(*dst, tag)
	}
}
