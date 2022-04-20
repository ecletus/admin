package admin

import (
	"github.com/ecletus/core"
	"github.com/go-aorm/aorm"
)

type Finder struct {
	Result                interface{}
	Unlimited             bool
	Limit                 func(s *Searcher) (db *aorm.DB)
	Count                 func(s *Searcher) (total int, err error)
	FindMany, FindOne     func(s *Searcher) (interface{}, error)
	RequestParserDisabled bool
}

// ParseFindMany parse context and find many records based on current conditions
func (this *Searcher) ParseFindMany(finder ...*Finder) (_ interface{}, err error) {
	if err = this.ParseContext(finder...); err != nil {
		if aorm.IsParseIdError(err) {
			return nil, nil
		}
		return
	}

	if this.HasError() {
		err = this.Errors
		this.Errors = core.Errors{}
		return
	}
	return this.Finder.FindMany(this)
}

// FindMany find many records based on current conditions
func (this *Searcher) FindMany() (_ interface{}, err error) {
	return this.Finder.FindMany(this)
}

func (this *Searcher) One(cb func()) {
	this.one = true
	defer func() {
		this.one = false
	}()
	cb()
}

// FindOne find one record based on current conditions
func (this *Searcher) FindOne(finder ...*Finder) (result interface{}, err error) {
	this.One(func() {
		if err = this.ParseContext(finder...); err != nil {
			return
		}
		if this.HasError() {
			err = this.Errors
			this.Errors = core.Errors{}
			return
		}
		result, err = this.Finder.FindOne(this)
	})
	return
}
