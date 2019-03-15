package admin

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/ecletus/core"
	"github.com/ecletus/core/utils"
	"github.com/ecletus/db/inheritance"
	"github.com/moisespsena-go/aorm"
)

type ChildMeta struct {
	Name            string
	Valuer          ChildMetaValuer
	FormattedValuer ChildMetaValuer
}

// GetFormattedValuer get formatted valuer from meta
func (meta *ChildMeta) GetFormattedValuer() ChildMetaValuer {
	if meta.FormattedValuer != nil {
		return meta.FormattedValuer
	}
	return meta.Valuer
}

type ChildOptions struct {
	Data map[interface{}]interface{}
}
type ChildMetaValuer func(record interface{}, context *core.Context, original MetaValuer) interface{}

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
	field     *aorm.Field
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
	if r.field, ok = r.Resource.FakeScope.FieldByName(r.FieldName); ok {
		r.dbName = "_ref_" + strconv.Itoa(index)
		rTbName := r.Resource.FakeScope.QuotedTableName()
		rFieldName := r.field.DBName
		pk := r.Resource.FakeScope.PrimaryKey()
		r.query = "(SELECT " + r.dbName + "." + pk + " FROM " + rTbName + " " + r.dbName + " WHERE " + r.dbName + "." + rFieldName + "_id = {} LIMIT 1) as " + r.dbName
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

func (r *Inheritance) Field() *aorm.Field {
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
	tbName := rs.resource.FakeScope.QuotedTableName()
	fName := rs.resource.FakeScope.PrimaryKey()
	tfname := tbName + "." + fName

	rs.columns = make([]string, len(rs.Items))

	for i, r := range rs.Items {
		rs.columns[i] = strings.Replace(r.query, "{}", tfname, -1)
	}

	rs.query = "SELECT " + strings.Join(rs.columns, ", ") + " FROM " + tbName + " WHERE " + tfname + " = ?"
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
	for i, _ := range r {
		r[i] = sql.NullString{}
	}
	return r
}
func (rs *Inheritances) Find(pk interface{}, db *aorm.DB) (r *Inheritance, err error) {
	if rs.query == "" {
		rs.Build()
	}
	db = db.Raw(rs.query, pk)
	if db.Error != nil {
		return r, db.Error
	}
	row := db.Row()
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

func (res *Resource) Inherit(super *Resource, fieldName string, options ...*ChildOptions) *Inheritance {
	r := super.Children.Add(&Inheritance{Resource: res, FieldName: fieldName})
	if len(options) == 0 || options[0] == nil {
		r.Options = &ChildOptions{}
	} else {
		r.Options = options[0]
	}
	res.Inherits[fieldName] = &Child{r, super}
	return r
}

func (res *Resource) GetChildMeta(record interface{}, fieldName string) *ChildMeta {
	if record != nil {
		r := record.(inheritance.ParentModelInterface)
		if child := r.GetQorChild(); child != nil {
			ref := res.Children.Items[child.Index]
			return ref.Options.GetMeta(fieldName)
		}
	}
	return nil
}

func (res *Resource) SetInheritedMeta(meta *Meta) *Meta {
	name := meta.Name
	meta.Name += "_inherited"
	if meta.DefaultLabel == "" {
		meta.DefaultLabel = utils.HumanizeString(name)
	}
	meta.Valuer = func(i interface{}, context *core.Context) interface{} {
		originalMeta := res.GetDefinedMeta(name).Valuer
		if meta := res.GetChildMeta(i, name); meta != nil {
			valuer := meta.GetFormattedValuer()
			return valuer(i, context, originalMeta)
		}
		if originalMeta != nil {
			return originalMeta(i, context)
		}
		return nil
	}
	meta.FormattedValuer = func(i interface{}, context *core.Context) interface{} {
		originalMeta := res.GetDefinedMeta(name).GetFormattedValuer()
		if meta := res.GetChildMeta(i, name); meta != nil {
			valuer := meta.GetFormattedValuer()
			return valuer(i, context, originalMeta)
		}
		return originalMeta(i, context)
	}
	res.Meta(&Meta{Name: name, Label: meta.Label})
	return res.SetMeta(meta)
}
