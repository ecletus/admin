package admin

import "github.com/moisespsena-go/aorm"

type CountAggregations []*CountAggregation

func (this *CountAggregations) Add(item ...*CountAggregation) {
	*this = append(*this, item...)
}

type CountAggregation struct {
	Name         string
	Resource     *Resource
	Record       interface{}
	Aggregations []*AggregationClause
}

type AggregationClause struct {
	Query     *aorm.Query
	QueryFunc aorm.QueryFunc
	FieldName string
	Embed     bool
}

type CountAggregationScopes map[string]*CountAggregations

func (this *CountAggregationScopes) Of(name string) *CountAggregations {
	if *this == nil {
		*this = CountAggregationScopes{}
	}
	if aggs := (*this)[name]; aggs != nil {
		return aggs
	}
	aggs := CountAggregations{}
	(*this)[name] = &aggs
	return &aggs
}

func (this CountAggregationScopes) Get(name string) CountAggregations {
	if this == nil {
		return nil
	}
	if aggs, ok := this[name]; ok {
		return *aggs
	}
	return nil
}
