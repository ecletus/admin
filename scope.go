package admin

import (
	"strings"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/ecletus/core/utils"
	"github.com/moisespsena-go/aorm"
)

// ScopeGroup register scopes into group for resource
func (this *Scheme) ScopeGroup(groupName string, scope ...*Scope) {
	this.ScopeGroupConfig(&ScopeConfig{Group: groupName}, scope...)
}

// Config register scopes into group for resource
func (this *Scheme) ScopeGroupConfig(cfg *ScopeConfig, scope ...*Scope) {
	for _, scope := range scope {
		if scope.ScopeConfig != nil {
			cfg.Wrap(scope.ScopeConfig)
		} else {
			scope.ScopeConfig = cfg
		}
		this.Scope(scope)
	}
}

type ScopeConfig struct {
	AdvancedFunc  func(ctx *Context) bool
	advancedFuncs []func(ctx *Context) bool

	Group          string
	GroupLabelFunc func(ctx *Context) string
}

func (this *ScopeConfig) Wrap(sub *ScopeConfig) {
	sub.Group = this.Group
	sub.GroupLabelFunc = this.GroupLabelFunc
	sub.advancedFuncs = append(sub.advancedFuncs, this.advancedFuncs...)
	if this.AdvancedFunc != nil {
		sub.advancedFuncs = append(sub.advancedFuncs, this.AdvancedFunc)
	}
}

func (this *ScopeConfig) Advanced(ctx *Context) bool {
	for _, f := range this.advancedFuncs {
		if f(ctx) {
			return true
		}
	}
	if this.AdvancedFunc != nil {
		return this.AdvancedFunc(ctx)
	}
	return false
}

// Scope scope definiation
type Scope struct {
	Name         string
	Label        string
	LabelFunc    func(ctx *Context) string
	Visible      func(context *Context) bool
	Handler      func(*aorm.DB, *Searcher, *core.Context) *aorm.DB
	Default      bool
	BaseResource *Resource
	*ScopeConfig
}

func (this *Scope) GetLabel(ctx *Context) (s string) {
	if this.LabelFunc != nil {
		if s = this.LabelFunc(ctx); s != "" {
			return
		}
	}

	var key = this.BaseResource.I18nPrefix + ".scopes."
	if this.Group != "" {
		key += this.Group + "."
	}
	key += this.Label

	return ctx.I18nT(key).Default(this.Label).Get()
}

func (this *Scope) GetGroupLabel(ctx *Context) (s string) {
	if this.GroupLabelFunc != nil {
		if s = this.GroupLabelFunc(ctx); s != "" {
			return
		}
	}
	return ctx.I18nT(this.BaseResource.I18nPrefix + ".scopes." + this.Group + ".label").Default(this.Group).Get()
}

func NewScope(name string, label string, handler func(*aorm.DB, *Searcher, *core.Context) *aorm.DB, defaul ...bool) *Scope {
	var d bool
	for _, d = range defaul {
	}
	return &Scope{Name: name, Label: label, Handler: handler, Default: d}
}

type SelectOneMetaNameToScopeGroupOptions struct {
	Operator       func(column string, value interface{}) aorm.Query
	DefaultValue   string
	DefaultValues  []string
	DefaultAlone   bool
	DefaultHandler func(db *aorm.DB, searcher *Searcher, context *core.Context) *aorm.DB
	All            bool
	AllLabel       string
	AllLabelFunc   func(ctx *Context) string
}

func ConvertSelectOneMetaNameToScopeGroup(res *Resource, metaName string) {
	ConvertSelectOneMetaNameToScopeGroupOpt(res, metaName, nil)
}

func ConvertSelectOneMetaNameToScopeGroupOpt(res *Resource, metaName string, opt *SelectOneMetaNameToScopeGroupOptions) {
	if opt == nil {
		opt = &SelectOneMetaNameToScopeGroupOptions{}
	}
	if opt.Operator == nil {
		opt.Operator = func(column string, value interface{}) aorm.Query {
			return *aorm.NewQuery("_."+column+" = ?", value)
		}
	}

	var (
		meta      = res.GetMetaOrSetDefault(&Meta{Name: metaName, Type: "select_one"})
		items     = meta.Config.(*SelectOneConfig).getCollection(nil, nil)
		defaults  [][2]string
		isDefault = func(name string) bool {
			for _, v := range opt.DefaultValues {
				if v == name {
					return true
				}
			}
			return false
		}
	)

	for i, pair := range items {
		items[i] = append(pair, utils.ToParamString(metaName)+"_"+pair[0])
		if isDefault(pair[0]) {
			if opt.DefaultAlone {
				defaults = append(defaults, [2]string{items[i][2] + "__default", pair[0]})
			} else {
				defaults = append(defaults, [2]string{items[i][2], pair[0]})
			}
		}
	}

	// TODO: Default Alone
	// if opt.DefaultAlone && len(defaults) > 0 {
	//	items = append(items, []string{opt.DefaultValue, defaults[0], defaults[0]})
	// }

	cfg := &ScopeConfig{
		Group:          metaName,
		GroupLabelFunc: meta.GetLabel,
	}

	for _, pair := range items {
		func(value, label, sname string) {
			var mv = &resource.MetaValue{
				Parent: &resource.MetaValue{},
				Name:   metaName,
				Meta:   meta,
				Value:  []string{value},
			}
			res.Scope(&Scope{
				ScopeConfig: cfg,
				Name:        sname,
				LabelFunc: func(ctx *Context) string {
					return label
				},
				Default: isDefault(sname) && len(defaults) == 1,
				Handler: func(db *aorm.DB, searcher *Searcher, context *core.Context) *aorm.DB {
					record := meta.BaseResource.New()
					dec := resource.DecodeToResource(
						meta.BaseResource,
						record,
						&resource.MetaValue{
							Name: meta.Name,
							MetaValues: &resource.MetaValues{
								Values: []*resource.MetaValue{mv},
								ByName: map[string]*resource.MetaValue{metaName: mv},
							},
						}, context, resource.ProcSkipLoad).Internal()

					err := dec.Start()
					if err != nil {
						context.AddError(err)
						return db
					}
					value := meta.Value(context, record)
					q := opt.Operator(meta.FieldStruct.DBName, value)
					return db.Where(q.Query, q.Args...)
				},
			})
		}(pair[0], pair[1], pair[2])
	}

	if len(defaults) > 0 {
		if len(defaults) > 1 {
			res.Scope(&Scope{
				Name:        "!default",
				ScopeConfig: cfg,
				Default:     true,
				Handler: func(db *aorm.DB, searcher *Searcher, context *core.Context) *aorm.DB {
					var q aorm.Query
					for _, defaul := range defaults {
						q2 := opt.Operator(meta.FieldStruct.DBName, defaul[1])
						q.Query += q2.Query + " OR "
						q.AddArgs(q2.Args...)
					}
					q.Query = strings.TrimSuffix(q.Query, " OR ")
					return db.Where(q.Query, q.Args...)
				},
			})
		}
	} else if opt.DefaultHandler != nil {
		res.Scope(&Scope{
			Name:        "!default",
			ScopeConfig: cfg,
			Default:     true,
			Handler:     opt.DefaultHandler,
		})
	}

	if opt.AllLabelFunc == nil && opt.AllLabel != "" {
		opt.AllLabelFunc = func(ctx *Context) string {
			return opt.AllLabel
		}
	}

	if opt.AllLabelFunc != nil {
		opt.All = true
	} else if opt.All {
		opt.AllLabelFunc = func(ctx *Context) string {
			return ctx.Ts(I18NGROUP + ".common.all")
		}
	}

	if opt.All {
		res.Scope(&Scope{
			Name:        utils.ToParamString(metaName) + "__all",
			ScopeConfig: cfg,
			LabelFunc:   opt.AllLabelFunc,
			Handler: func(db *aorm.DB, searcher *Searcher, context *core.Context) *aorm.DB {
				return db
			},
		})
	}
}
