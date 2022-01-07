package admin

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/ecletus/core"
	"github.com/ecletus/db/inheritance"
	"github.com/moisespsena-go/aorm"
)

type ChildMeta struct {
	Name            string
	Valuer          ChildMetaValuer
	FormattedValuer ChildMetaFValuer
}

// GetFormattedValuer get formatted valuer from meta
func (meta *ChildMeta) GetFormattedValuer() ChildMetaFValuer {
	if meta.FormattedValuer != nil {
		return meta.FormattedValuer
	}
	return func(record interface{}, context *core.Context, original MetaFValuer) *FormattedValue {
		return &FormattedValue{Record: record, Raw: meta.Valuer(record, context, func(record interface{}, context *core.Context) interface{} {
			v := original(record, context)
			if v == nil {
				return nil
			}
			return v.Raw
		})}
	}
}

type ChildOptions struct {
	Data map[interface{}]interface{}
}
type ChildMetaValuer func(record interface{}, context *core.Context, original MetaValuer) interface{}
type ChildMetaFValuer func(record interface{}, context *core.Context, original MetaFValuer) *FormattedValue

func (ro *ChildOptions) Meta(meta *ChildMeta) *ChildOptions {
	if ro.Data == nil {
		ro.Data = make(map[interface{}]interface{})
	}
	ro.Data["meta."+meta.Name] = meta
	return ro
}

func (ro *ChildOptions) GetMeta(metaName string) *ChildMeta {
	if v, ok := ro.Data["meta."+metaName]; ok {
		return v.(*ChildMeta)
	}
	return nil
}

type Inheritance struct {
	Resource  *Resource
	FieldName string
	index     int
	dbName    string
	field     *aorm.StructField
	query     string
	Options   *ChildOptions
}

type Child struct {
	Inheritance *Inheritance
	Resource    *Resource
}

func (r *Inheritance) build(index int) {
	r.index = index
	var ok bool
	if r.field, ok = r.Resource.ModelStruct.FieldsByName[r.FieldName]; ok {
		r.dbName = "_ref_" + strconv.Itoa(index)
		rFieldName := r.field.DBName
		pk := r.Resource.ModelStruct.PrimaryField().DBName
		r.query = "(SELECT " + r.dbName + "." + pk + " FROM / " + r.dbName + " WHERE " + r.dbName + "." + rFieldName + "_id = {} LIMIT 1) as " + r.dbName
	} else {
		panic(fmt.Errorf("Invalid field %q", r.FieldName))
	}
}

func (r *Inheritance) Index() int {
	return r.index
}

func (r *Inheritance) DBName() string {
	return r.dbName
}

func (r *Inheritance) Field() *aorm.StructField {
	return r.field
}

type Inheritances struct {
	Items    []*Inheritance
	query    string
	resource *Resource
	lock     sync.Mutex
	columns  []string
}

func (rs *Inheritances) Add(r *Inheritance) *Inheritance {
	if rs.query != "" {
		panic(fmt.Errorf("Inheritances for %v Resource has be builded.", rs.resource.Param))
	}
	r.build(len(rs.Items))
	rs.Items = append(rs.Items, r)
	return r
}

func (rs *Inheritances) Build() {
	if rs.query != "" {
		panic(fmt.Errorf("Inheritances for %v Resource has be builded.", rs.resource.Param))
	}
	rs.lock.Lock()
	defer rs.lock.Unlock()
	if rs.query != "" {
		return
	}
	fName := rs.resource.ModelStruct.PrimaryField().DBName
	tfname := "/." + fName

	rs.columns = make([]string, len(rs.Items))

	for i, r := range rs.Items {
		rs.columns[i] = strings.Replace(r.query, "{}", tfname, -1)
	}

	rs.query = "SELECT " + strings.Join(rs.columns, ", ") + " FROM / WHERE " + tfname + " = ?"
}

func (rs *Inheritances) Query() string {
	return rs.query
}

func (rs *Inheritances) Columns() []string {
	if rs.query == "" {
		rs.Build()
	}
	return rs.columns
}
func (rs *Inheritances) NewSlice() []interface{} {
	r := make([]interface{}, len(rs.Items))
	for i := range r {
		r[i] = sql.NullString{}
	}
	return r
}
func (rs *Inheritances) Find(pk interface{}, db *aorm.DB) (r *Inheritance, err error) {
	if rs.query == "" {
		rs.Build()
	}
	query := strings.ReplaceAll(rs.query, "/", rs.resource.QuotedTableName(db))
	db = db.Raw(query, pk)
	if db.Error != nil {
		return r, db.Error
	}
	var row *sql.Row
	if row, err = db.Row(); err != nil {
		return
	}
	results := make([]interface{}, len(rs.Items))
	err = row.Scan(&results)
	if err != nil {
		return
	}
	for i, v := range results {
		if v != nil {
			return rs.Items[i], nil
		}
	}
	return
}

func (this *Resource) Inherit(super *Resource, fieldName string, options ...*ChildOptions) *Inheritance {
	r := super.Children.Add(&Inheritance{Resource: this, FieldName: fieldName})
	if len(options) == 0 || options[0] == nil {
		r.Options = &ChildOptions{}
	} else {
		r.Options = options[0]
	}
	this.Inherits[fieldName] = &Child{r, super}
	return r
}

func (this *Resource) GetChildMeta(record interface{}, fieldName string) *ChildMeta {
	if record != nil {
		r := record.(inheritance.ParentModelInterface)
		if child := r.GetQorChild(); child != nil {
			ref := this.Children.Items[child.Index]
			return ref.Options.GetMeta(fieldName)
		}
	}
	return nil
}

func (this *Resource) SetInheritedMeta(meta *Meta) *Meta {
	panic("not implemented")
	// todo: check formatted values ZeroFunc
	/* name := meta.Name
	meta.Name += "_inherited"
	if meta.DefaultLabel == "" {
		meta.DefaultLabel = utils.HumanizeString(name)
	}
	meta.Valuer = func(i interface{}, context *core.Context) interface{} {
		originalMeta := this.GetDefinedMeta(name).Valuer
		if meta := this.GetChildMeta(i, name); meta != nil {
			valuer := meta.GetFormattedValuer()
			return valuer(i, context, func(record interface{}, context *core.Context) *FormattedValue {
				v := originalMeta(record, context)
				if v == nil {
					return nil
				}
				return &FormattedValue{Raw: v}
			})
		}
		if originalMeta != nil {
			return originalMeta(i, context)
		}
		return nil
	}
	meta.FormattedValuer = func(i interface{}, context *core.Context) *FormattedValue {
		original := this.GetDefinedMeta(name).GetFormattedValuer()
		if meta := this.GetChildMeta(i, name); meta != nil {
			valuer := meta.GetFormattedValuer()
			return valuer(i, context, original)
		}
		return original(i, context)
	}
	this.Meta(&Meta{Name: name, Label: meta.Label})
	return this.SetMeta(meta)
	*/
}
