package admin

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"reflect"
	"strconv"
	"strings"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/moisespsena-go/tracederror"
	"github.com/moisespsena/template/funcs"
	"github.com/moisespsena/template/html/template"
	"github.com/pkg/errors"
)

func (this *Context) renderForm(state *template.State, value interface{}, sections []*Section) {
	prefix := this.FormOptions.InputPrefix
	if prefix == "" {
		prefix = resource.DefaultFormInputPrefix
	}
	if this.MetaStack.Empty() {
		defer this.MetaStack.Push(&Meta{Name: prefix})()
	}
	this.renderSections(state, value, []string{prefix}, state.Writer(), "form", this.Type.Has(SHOW) || this.Type.Has(INDEX), sections...)
}

func (this *Context) renderSections(state *template.State, value interface{}, prefix []string, writer io.Writer, kind string, readOnly bool, sections ...*Section) {
	var (
		res            *Resource
		getMeta        func(string) *Meta
		buf            bytes.Buffer
		rendered       = map[string]bool{}
		skipAttrsCheck = map[string]bool{}
		add            = func(dst *[]template.HTML) {
			if buf.Len() == 0 {
				return
			}
			s := strings.TrimSpace(string(buf.Bytes()))
			buf.Reset()
			if s != "" {
				*dst = append(*dst, template.HTML(s))
			}
		}
	)

	if reflect.TypeOf(value).Kind() != reflect.Ptr {
		value = func(v interface{}) interface{} {
			nv := reflect.New(reflect.TypeOf(v))
			nv.Elem().Set(reflect.ValueOf(v))
			return nv.Elem().Addr().Interface()
		}(value)
	}

	for _, section := range sections {
		var (
			hasRequired bool
			rows        [][]template.HTML
		)

		if res != section.Resource {
			res = section.Resource
			getMeta = section.Resource.MetaContextGetter(this)
		}

		for i := 0; i < len(section.Rows); i++ {
			var (
				column      = section.Rows[i]
				exclude     int
				columnsHtml []template.HTML
			)

		colsLoop:
			for j := 0; j < len(column); j++ {
				var col = column[j]
				if col == "" || col == nil {
					continue
				}
				switch col := col.(type) {
				case string:
					switch col {
					case ":SUBMIT":
						meta := &Meta{
							BaseResource: res,
							Type:         "submit",
							Meta: &resource.Meta{
								BaseResource: res,
								MetaName:     &resource.MetaName{Name: ":SUBMIT"},
								Valuer: func(record interface{}, context *core.Context) interface{} {
									if res.Config.SubmitLabel == "" {
										key := res.Config.SubmitLabelKey
										if key == "" {
											key = I18NGROUP + ".form.submit"
										}
										return context.Ts(key, "Submit")
									}
									return res.Config.SubmitLabel
								},
							},
						}
						this.renderMeta(state, meta, value, prefix, kind, "", &buf)
						rendered[col] = true
						add(&columnsHtml)
						continue colsLoop
					}
					meta := getMeta(col)
					if meta != nil {
						if meta.IsEnabled(value, this, meta, readOnly) {
							if meta.IsRequired() {
								hasRequired = true
							}
							_, skipAttrCheck := skipAttrsCheck[col]
							if attS := meta.Tags.GetString("ATTR"); attS != "" && !skipAttrCheck {
								skipAttrsCheck[col] = true

								if attS[0] == ';' {
									// append to new sections
									var news []interface{}
									for _, col2 := range strings.Split(attS, ";")[1:] {
										col2 = strings.TrimSpace(col2)
										if col2 == "." {
											col2 = col
										}
										news = append(news, col2)
									}
									news = append(news, column[j+1:]...)
									column = append(column[0:j], news...)
									section.Rows[i] = column
									section.Rows = append(section.Rows[0:i], append([][]interface{}{news}, section.Rows[i+1:]...)...)
									j--
								} else {
									for _, col2 := range strings.Split(attS, ";") {
										col2 = strings.TrimSpace(col2)
										if col2 == "." {
											this.renderMeta(state, meta, value, prefix, kind, "", &buf)
											add(&columnsHtml)
										} else {
											m := getMeta(col2)
											if m == nil {
												panic(fmt.Errorf("Resource %q: meta %s: meta %q in TAG[ATTR]=%q is nil", this.Resource.ID, col, col2, attS))
											}
											this.renderMeta(state, m, value, prefix, kind, "", &buf)
											add(&columnsHtml)
											rendered[col2] = true
										}
									}
								}
							} else {
								this.renderMeta(state, meta, value, prefix, kind, "", &buf)
								rendered[col] = true
								add(&columnsHtml)
							}
						} else {
							exclude++
						}
					}

					if _, ok := rendered[col]; ok {
						continue colsLoop
					}
				case *Section:
					this.renderSections(state, value, prefix, &buf, kind, readOnly, col)
					add(&columnsHtml)
				}

			}

			if !hasRequired && (this.Action == "show" || this.Action == "action_show") && len(columnsHtml) == 0 {
				continue
			}

			rows = append(rows, columnsHtml)
		}

		if len(rows) > 0 {
			var data = map[string]interface{}{
				"Section":  section,
				"Title":    template.HTML(section.Title),
				"Rows":     rows,
				"ReadOnly": readOnly,
			}

			if executor, err := this.GetTemplate("metas/section"); err == nil {
				err = executor.Execute(writer, data, this.FuncValues())
			}
		}
	}
}

func (this *Context) renderFilter(filter *Filter) template.HTML {
	var (
		err      error
		executor *template.Executor
		dir      = "filter"
		result   = bytes.NewBufferString("")
		advanced = filter.IsAdvanced()
	)

	if advanced {
		dir = "advanced_filter"
	}

	if executor, err = this.GetTemplate(fmt.Sprintf("metas/%v/%v", dir, filter.Type)); err == nil {
		var label, prefix string
		if !filter.LabelDisabled {
			label = filter.GetLabelC(this.Context)
		}
		if advanced {
			prefix = "adv_"
		}

		var data = map[string]interface{}{
			"Filter":          filter,
			"Label":           label,
			"InputNamePrefix": fmt.Sprintf("%sfilter[%v]", prefix, filter.Name),
			"Context":         this,
			"Resource":        this.Resource,
			"Arg":             this.Searcher.filters[filter],
		}

		err = executor.Execute(result, data, this.FuncValues())
	}

	if err != nil {
		this.AddError(err)
		result.WriteString(errors.Wrap(err, fmt.Sprintf("render filter template for %v(%v)", filter.Name, filter.Type)).Error())
	}

	return template.HTML(result.String())
}

func (this *Context) savedFilters() (filters []SavedFilter) {
	this.Admin.settings.Get("saved_filters", &filters, this)
	return
}

func (this *Context) NestedForm() bool {
	return this.nestedForm > 0
}

func (this *Context) renderMetaWithPath(state *template.State, pth string, record interface{}, meta *Meta, types ...string) {
	var (
		typ   = "index"
		mode  string
		parts = strings.Split(pth, "/")
	)

	defer this.MetaStack.PushNames(parts...)()

	for _, t := range types {
		if strings.HasPrefix(t, "mode-") {
			mode = strings.TrimPrefix(t, "mode-")
		} else if t != "" {
			typ = t
		}
	}

	this.renderMeta(state, meta, record, this.MetaStack.Path(), typ, mode, NewTrimLeftWriter(state.Writer()))
}

func (this *Context) renderMeta(state *template.State, meta *Meta, record interface{}, prefix []string, kind, mode string, writer io.Writer) {
	defer this.MetaStack.Push(meta)()

	if mode == "single" {
		oldFlags := this.RenderFlags
		this.RenderFlags |= CtxRenderMetaSingleMode
		this.Type |= INLINE
		defer func() {
			this.RenderFlags = oldFlags
		}()
	}

	var (
		err             error
		funcsMap        = funcs.FuncMap{}
		executor        *template.Executor
		show            = this.Type.Has(PRINT, SHOW, INDEX) || kind == "index" || kind == "show"
		nestedFormCount int
		readOnly        = show
		fv              *FormattedValue
		value           interface{}
		readOnlyCalled  bool

		mctx = &MetaContext{
			Meta:     meta,
			Out:      writer,
			Context:  this,
			Record:   record,
			Prefix:   prefix,
			ReadOnly: readOnly,
		}
	)
	defer func() {
		for _, cb := range mctx.deferRenderHandlers {
			cb()
		}
	}()

	meta.CallPrepareContextHandlers(mctx, record)
	meta.CallBeforeRenderHandlers(mctx, record)
	readOnly = mctx.ReadOnly

	if !readOnly {
		readOnlyCalled = true
		if readOnly = meta.IsReadOnly(mctx.Context, record); readOnly {
			var newType = SHOW
			if mctx.Context == this {
				clone := *this
				clone.Type = this.Type.ClearCrud().Set(newType)
				mctx.Context = &clone
			} else {
				mctx.Context.Type = mctx.Context.Type.ClearCrud().Set(newType)
			}
		}
	}

	mctx.ReadOnly = readOnly
	fv = meta.GetFormattedValue(mctx.Context, record, readOnly)
	mctx.FormattedValue = fv

	if readOnly && mctx.Context.Type.Has(EDIT, NEW, SHOW) {
		if fv == nil {
			return
		}
		if fv.Raw != nil {
			switch fv.Value {
			case "":
				if fv.SafeValue == "" {
					return
				}
			}
		}
	}
	if show && !meta.IsRequired() {
		if fv == nil {
			return
		}
		if !meta.ForceShowZero {
			if fv.IsZero() {
				return
			} else if !meta.ForceEmptyFormattedRender {
				if fv.Raw == nil && fv.Raws == nil {
					return
				}

				if fv.Value == "" && fv.SafeValue == "" {
					return
				}
			}
		}
	}

	if !meta.Include {
		prefix = append(prefix, meta.Name)
	}

	var nestedRenderSectionsContext = func(state *template.State, kind string, this *Context, typ string, meta *Meta, index int, prefx ...string) {
		var (
			sections []*Section
			readOnly = readOnly
			record   = this.ResourceRecord
			ctyp     = ParseContextType(typ)
		)

		switch ctyp {
		case NEW:
			sections = this.newMetaSections(meta)
		case EDIT:
			sections = this.editMetaSections(meta)
		case SHOW:
			sections = this.showMetaSections(meta)
			readOnly = true
		case INDEX:
			sections = this.indexSections()
		}

		oldTemplateName, oldAction, oldTyp := this.TemplateName, this.Action, this.Type
		this.TemplateName = oldTyp.String()
		this.Action = oldTyp.String()
		this.SetBasicType(ctyp)

		defer func() {
			this.TemplateName, this.Action, this.Type = oldTemplateName, oldAction, oldTyp
		}()

		this.nestedForm++

		switch index {
		case -2:
			// defer this.MetaStack.Push(meta)()
		case -1:
			defer this.MetaStack.Push(meta, "{{index}}")()
		default:
			defer this.MetaStack.Push(meta, strconv.Itoa(index))()
		}

		if record == nil && !show && meta.Resource != nil {
			record = meta.Resource.New()
		}

		defer func() {
			nestedFormCount++
			this.nestedForm--
		}()

		newPrefix := append([]string{}, prefix...)

		if len(prefx) > 0 && prefx[0] != "" {
			for prefx[0][0] == '.' {
				newPrefix = newPrefix[0 : len(newPrefix)-1]
				prefx[0] = prefx[0][1:]
			}

			newPrefix = append(newPrefix, prefx...)
		}

		if index >= 0 {
			last := newPrefix[len(newPrefix)-1]
			newPrefix = append(newPrefix[:len(newPrefix)-1], last, strconv.Itoa(index))
		} else if index == -1 {
			last := newPrefix[len(newPrefix)-1]
			newPrefix = append(newPrefix[:len(newPrefix)-1], last, "{{index}}")
		}

		if len(sections) > 0 {
			w := NewTrimLeftWriter(state.Writer())
			this.renderSections(state, record, newPrefix, w, kind, readOnly, sections...)
		}
	}

	funcsMap["render_nested_ctx"] = nestedRenderSectionsContext

	defer func() {
		if err != nil {
			panic(err)
		}
		if r := recover(); r != nil {
			var (
				msg          string
				metaTreePath = path.Join(mctx.Context.MetaStack.Path()...)
			)
			msg = fmt.Sprintf("render meta %q (%v)", metaTreePath, kind)
			mctx.Out.Write([]byte(msg))

			if et, ok := r.(tracederror.TracedError); ok {
				panic(tracederror.Wrap(et, msg))
			} else if err, ok := r.(error); ok {
				panic(tracederror.New(errors.Wrap(err, msg), et.Trace()))
			} else {
				panic(tracederror.New(errors.Wrap(fmt.Errorf("recoverd_error %T: %v", r, r), msg)))
			}
		}
	}()

	var (
		others           []string
		typeTemplateName string
		h                = &MetaConfigHelper{this.MetaStack.AnyIndexPathString()}
	)

	if executor, err = h.GetTemplateExecutor(this, kind, meta, fv); err != nil {
		goto failed
	} else if executor == nil {
		if v := h.GetTemplate(&this.LocalContext); v != "" {
			if executor, err = mctx.Context.GetTemplateOrDefault(v,
				TemplateExecutorMetaValue, others...); err != nil {
				err = errors.Wrapf(err, "meta %v", strings.Join(mctx.Context.MetaStack.Path(), "."))
			}
		} else if typeTemplateName = h.GetTypeName(&this.LocalContext); typeTemplateName == "" && meta.Config != nil {
			switch cfg := meta.Config.(type) {
			case MetaTemplateExecutorGetter:
				if executor, err = cfg.GetTemplateExecutor(mctx.Context, record, kind, readOnly); err != nil {
					goto failed
				}
			case MetaTemplateNameGetter:
				var templateName string
				if templateName, err = cfg.GetTemplateName(mctx.Context, record, kind, readOnly); err != nil {
					goto failed
				}
				if templateName != "" {
					others = append(others, templateName)
				}
			case MetaUserTypeTemplateNameGetter:
				if typeTemplateName, err = cfg.GetUserTypeTemplateName(mctx.Context, record, readOnly); err != nil {
					goto failed
				}
			}
		}
	}

	if typeTemplateName == "" {
		typeTemplateName = meta.GetType(record, mctx.Context, readOnly)
	}

	if typeTemplateName != "" {
		others = append(others, fmt.Sprintf("metas/%v/%v", kind, typeTemplateName))
	}

	if executor, err = mctx.Context.GetTemplateOrDefault(fmt.Sprintf("%v/metas/%v/%v", meta.BaseResource.ToParam(), kind, meta.Name),
		TemplateExecutorMetaValue, others...); err != nil {
		err = errors.Wrapf(err, "meta %v", strings.Join(mctx.Context.MetaStack.Path(), "."))
	} else {
		parts := strings.SplitN(executor.Template().Path, "/metas/", 2)
		if len(parts) == 2 {
			kind = strings.TrimSuffix(parts[1], ".tmpl")
		}
	}

	if fv == nil {
		fv = &FormattedValue{Zero: true}
	}

	if fv.SafeValue != "" {
		value = template.HTML(fv.SafeValue)
	} else {
		value = fv.Value
	}

	if err == nil {
		if !readOnly && !readOnlyCalled {
			readOnly = meta.IsReadOnly(mctx.Context, record)
			mctx.ReadOnly = readOnly
		}
		var data = map[string]interface{}{
			"Context":         mctx.Context,
			"MetaType":        kind,
			"BaseResource":    meta.BaseResource,
			"Meta":            meta,
			"Record":          record,
			"ResourceValue":   record,
			"MetaValue":       fv,
			"Value":           value,
			"Label":           meta.Label,
			"InputName":       strings.Join(prefix, "."),
			"InputParentName": strings.Join(prefix[0:len(prefix)-1], "."),
			"ModeSingle":      mode == "single",
			"MetaHelper":      h,
		}
		data["ReloadValue"] = func() {
			fv = meta.GetFormattedValue(mctx.Context, record, mctx.ReadOnly)
			var value interface{}
			if fv.SafeValue != "" {
				value = template.HTML(fv.SafeValue)
			} else {
				value = fv.Value
			}
			data["MetaValue"] = fv
			data["Value"] = value
		}
		data["InputId"] = strings.ReplaceAll(strings.Join(mctx.Context.MetaStack.Path(), "_"), ".", "_")
		executor.SetSuper(state)
		mctx.Template = executor
		mctx.TemplateData = data
		meta.CallBeforeDoRenderHandler(mctx, record)
		data["ReadOnly"] = mctx.ReadOnly
		data["NotReadOnly"] = !mctx.ReadOnly
		err = executor.Execute(mctx.Out, data, mctx.Context.FuncValues(), funcsMap)
	}

failed:

	if err != nil {
		err = tracederror.TracedWrap(err, "got error when render meta %v template for %v(%v)", kind, meta.Name, meta.Type)
	}
}

type MetaContext struct {
	Meta                *Meta
	Out                 io.Writer
	Context             *Context
	Record              interface{}
	Prefix              []string
	ReadOnly            bool
	Template            *template.Executor
	TemplateData        map[string]interface{}
	FormattedValue      *FormattedValue
	deferRenderHandlers []func()
}

func (this *MetaContext) DeferHandler(f func()) {
	this.deferRenderHandlers = append(this.deferRenderHandlers, f)
}
