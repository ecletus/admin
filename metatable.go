package admin

import (
	"fmt"
	"os"
	"sort"
)

type MetaTableHeaders []*MetaTableHeader

func (this MetaTableHeaders) Sort() {
	var compare = func(x, y int) int {
		if x < y {
			return -1
		}
		if x == y {
			return 0
		}
		return 1
	}
	sort.Slice(this, func(i, j int) bool {
		var (
			a, b = this[i], this[j]
			pri  = compare(a.Row, b.Row)
			sec  = compare(a.Col, b.Col)
		)
		if pri != 0 {
			return pri < 0
		}
		return sec < 0
	})
}

type TreeSection struct {
	*Section
	Sections []*TreeSection
}

type MetaTableHeader struct {
	Label                             string
	Section                           *Section
	Meta                              *Meta
	Depth, Row, Col, RowSpan, ColSpan int
}

func (m MetaTableHeader) String() string {
	var s string
	if m.Section != nil {
		s = m.Section.Title
	}
	if m.Meta != nil {
		s = m.Meta.Name
	}
	return fmt.Sprintf("rs=%d, cs=%d: %s", m.RowSpan, m.ColSpan, s)
}

func (m MetaTableHeader) Tag() string {
	var s string
	if m.Section != nil {
		s += m.Section.Title
	}
	if m.Meta != nil {
		s += m.Meta.Name
	}
	s += m.Label
	return fmt.Sprintf("%s [%d %d %d]", s, m.Row, m.RowSpan, m.ColSpan)
}

type MetaTableMeta struct {
	Label    string
	Children []*MetaTableMeta
	Meta     *Meta
	Section  *Section
}

type MetasTable struct {
	Headers [][]*MetaTableHeader
	NumRows int
	Metas   []*Meta
}

func (t *MetasTable) RowTag() {
	var (
		print = func(arg ...interface{}) {
			fmt.Fprint(os.Stdout, arg...)
		}
		println = func(arg ...interface{}) {
			fmt.Fprintln(os.Stdout, arg...)
		}
	)
	for i, row := range t.Headers {
		print(i + 1)
		println(" ->")
		for i, col := range row {
			println("  -> ", i+1, col.Tag())
		}
		println()
	}
}

func ConvertSectionsToMetasTable(sections Sections, mg func(name string) *Meta) *MetasTable {
	var (
		mt      MetasTable
		toMetas func(s *Section) []*MetaTableMeta
		root    MetaTableMeta
	)

	toMetas = func(s *Section) (ret []*MetaTableMeta) {
		for _, row := range s.Rows {
			for _, col := range row {
				switch t := col.(type) {
				case *Section:
					if t.Title == "" {
						ret = append(ret, toMetas(t)...)
					} else {
						s2 := &MetaTableMeta{Section: t}
						s2.Children = toMetas(t)
						ret = append(ret, s2)
					}
				case string:
					m := mg(t)
					ret = append(ret, &MetaTableMeta{Meta: m})
					mt.Metas = append(mt.Metas, m)
				}
			}
		}
		return ret
	}

	for _, s := range sections {
		items := toMetas(s)
		if s.Title == "" {
			root.Children = append(root.Children, items...)
		} else {
			root.Children = append(root.Children, &MetaTableMeta{Section: s, Children: items})
		}
	}

	var (
		lcm = func(a, b int) int {
			c := a * b
			for b > 0 {
				t := b
				b = a % b
				a = t
			}
			return c / a
		}
		rowsToUse, width func(t *MetaTableMeta) int
		getCells         func(depth int, t *MetaTableMeta, row, col, rowsLeft int) []*MetaTableHeader
	)
	rowsToUse = func(t *MetaTableMeta) int {
		var childrenRows int
		if len(t.Children) > 0 {
			childrenRows++
		}

		for _, child := range t.Children {
			childrenRows = lcm(childrenRows, rowsToUse(child))
		}
		return 1 + childrenRows
	}
	width = func(t *MetaTableMeta) int {
		if len(t.Children) == 0 {
			return 1
		}
		w := 0
		for _, child := range t.Children {
			w += width(child)
		}
		return w
	}
	getCells = func(depth int, t *MetaTableMeta, row, col, rowsLeft int) (cells []*MetaTableHeader) {
		// Add top-most cell corresponding to the root of the current tree.
		rootRows := rowsLeft / rowsToUse(t)
		cells = append(cells, &MetaTableHeader{t.Label, t.Section, t.Meta, depth, row, col, rootRows, width(t)})
		for _, child := range t.Children {
			cells = append(cells, getCells(depth+1, child, row+rootRows, col, rowsLeft-rootRows)...)
			col += width(child)
		}
		if (row + 1) > mt.NumRows {
			mt.NumRows = row + 1
		}
		return
	}
	cells := getCells(0, &root, 0, 0, rowsToUse(&root))
	MetaTableHeaders(cells).Sort()
	mt.Headers = make([][]*MetaTableHeader, mt.NumRows)

	for _, cell := range cells {
		mt.Headers[cell.Row] = append(mt.Headers[cell.Row], cell)
	}
	mt.Headers = mt.Headers[1:]
	mt.NumRows--

	return &mt
}
