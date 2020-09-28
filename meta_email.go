package admin

import (
	"fmt"
	"reflect"
	"strings"

	"unapu.com/checkmail"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/moisespsena-go/aorm/types"
)

type EmailConfig struct {
	StringConfig
	SkipHostValidation,
	SkipUserValidation bool
}

func (e *EmailConfig) ConfigureQorMeta(metaor resource.Metaor) {
	m := metaor.(*Meta)
	if m.Type == "" {
		m.Type = "email"
	}
	e.StringConfig.MaxLen = types.EmailSize
	e.StringConfig.ConfigureQorMeta(metaor)
	m.Validator(func(record interface{}, values *resource.MetaValue, ctx *core.Context) (err error) {
		if v := values.FirstStringValue(); v != "" {
			v = strings.ToLower(v)
			values.Value.([]string)[0] = v
			doErr := resource.ErrField(ctx, record, values.Meta.GetFieldName(), values.Meta.GetRecordLabelC(ctx, record))
			if checkmail.ValidateFormat(v) != nil {
				return doErr(fmt.Sprintf(ctx.Ts(I18NGROUP+".errors.validations.email.bad_format"), v))
			}

			if !e.SkipUserValidation {
				if err := checkmail.ValidateHost(v); err != nil {
					if smtpErr, ok := err.(checkmail.SmtpError); ok {
						if smtpErr.Code() == "550" {
							return doErr(fmt.Sprintf(ctx.Ts(I18NGROUP+".errors.validations.email.account_not_exists"), v))
						}
					} else {
						return doErr(fmt.Sprintf(ctx.Ts(I18NGROUP+".errors.validations.email.host_not_exists"), v))
					}
				}
			} else if !e.SkipHostValidation {
				if err := checkmail.ValidateHost(v); err != nil {
					return doErr(fmt.Sprintf(ctx.Ts(I18NGROUP+".errors.validations.email.host_not_exists"), v))
				}
			}
		}
		return
	})
}

func init() {
	cfg := func(meta *Meta) {
		if meta.Config == nil {
			cfg := &EmailConfig{}
			meta.Config = cfg
			cfg.ConfigureQorMeta(meta)
		} else {
			meta.Config.ConfigureQorMeta(meta)
		}
	}
	RegisterMetaConfigor("email", cfg)
	RegisterMetaTypeConfigor(reflect.TypeOf(types.Email("")), cfg)
}
