package admin

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/ecletus/core"
	"github.com/ecletus/core/utils"
	"github.com/go-aorm/aorm"
)

type filterField struct {
	Field     *FieldFilter
	Operation string
	Typ       reflect.Type
}

func (f filterField) Apply(arg interface{}) (query string, argx interface{}) {
	if f.Field.Applyer != nil {
		return f.Field.Applyer(f.Field.FormatTerm(arg))
	}

	op := f.Operation
	fieldName := f.Field.QueryField()

	var cb = func() string {
		return fieldName + " " + op + " ?"
	}

	if b, ok := arg.(bool); ok {
		if b {
			return fieldName, nil
		} else {
			return "NOT " + fieldName, nil
		}
	}

	if f.Field.Struct.Struct.Type.Kind() == reflect.String {
		cb = func() string {
			return "UPPER(" + fieldName + ") " + op + " UPPER(?)"
		}
	} else if op == "" {
		op = "eq"
	}
	switch op {
	case "=", "eq", "equal":
		op = "="
	case "!", "ne":
		op = "!="
	case "btw":
		op = "BETWEEN"
		cb = func() string {
			return fieldName + " BETWEEN ? AND ?"
		}
	case ">", "gt":
		op = ">"
	case "<", "lt":
		op = "<"
	default:
		if f.Field.Struct.Struct.Type.Kind() == reflect.String {
			op = "LIKE"
			arg = "%" + arg.(string) + "%"
		}
	}
	return cb(), f.Field.FormatTerm(arg)
}

func filterResourceByFields(res *Resource, filterFields []*filterField, keyword string, db *aorm.DB, context *core.Context) *aorm.DB {
	var (
		joinConditionsMap  = map[string][]string{}
		conditions         []string
		keywords           []interface{}
		wheres             []interface{}
		generateConditions func(field *filterField, scope *aorm.Scope)
	)

	generateConditions = func(ff *filterField, scope *aorm.Scope) {
		field := ff.Field.Struct

		apply := func(kw interface{}) {
			query, arg := ff.Apply(kw)
			if query != "" {
				conditions = append(conditions, query)
				if arg != nil {
					keywords = append(keywords, arg)
				}
			} else if arg != nil {
				wheres = append(wheres, arg.(aorm.WhereClauser))
			}
		}

		appendString := func() {
			apply(keyword)
		}

		appendInteger := func() {
			if _, err := strconv.Atoi(keyword); err == nil {
				apply(keyword)
			}
		}

		appendFloat := func() {
			if _, err := strconv.ParseFloat(keyword, 64); err == nil {
				apply(keyword)
			}
		}

		appendBool := func() {
			if value, err := strconv.ParseBool(keyword); err == nil {
				apply(value)
			}
		}

		appendTime := func() {
			if parsedTime, err := utils.ParseTime(keyword, context); err == nil {
				apply(parsedTime)
			}
		}

		appendStruct := func() {
			v := reflect.New(field.Struct.Type)
			switch v.Elem().Interface().(type) {
			case time.Time, *time.Time:
				appendTime()
			case sql.NullInt64:
				appendInteger()
			case sql.NullFloat64:
				appendFloat()
			case sql.NullString:
				appendString()
			case sql.NullBool:
				appendBool()
			default:
				switch t := v.Interface().(type) {
				case aorm.StringParser:
					t.ParseString(keyword)
					apply(t)
				}
				// if we don't recognize the struct type, just ignore it
			}
		}

		switch indirectType(field.Struct.Type).Kind() {
		case reflect.String:
			appendString()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			appendInteger()
		case reflect.Float32, reflect.Float64:
			appendFloat()
		case reflect.Bool:
			appendBool()
		case reflect.Struct, reflect.Ptr:
			appendStruct()
		default:
			apply(keyword)
		}

		if ff.Field.InlineQueryName[0] == '{' {
			// has inline preload
			if preloadName := ff.Field.InlineQueryName[1:strings.IndexByte(ff.Field.InlineQueryName, '}')]; preloadName != "" {
				db = db.InlinePreload(preloadName)
			}
		}
	}

	scope := db.NewModelScope(res.ModelStruct, res.Value)
	for _, field := range filterFields {
		generateConditions(field, scope)
	}

	// join conditions
	if len(joinConditionsMap) > 0 {
		var joinConditions []string
		for key, values := range joinConditionsMap {
			joinConditions = append(joinConditions, fmt.Sprintf("%v %v", key, strings.Join(values, " AND ")))
		}
		db = db.Joins(strings.Join(joinConditions, " "))
	}

	// search conditions
	if len(conditions) > 0 {
		db = db.Where(aorm.IQ(strings.Join(conditions, " OR ")), keywords...)
	}

	for _, c := range wheres {
		db = db.Where(c)
	}

	return db
}
