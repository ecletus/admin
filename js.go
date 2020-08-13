package admin

import "html/template"

type JS struct {
	Data template.JS
	Raw  bool
}

func (this JS) String() string {
	return string(this.Data)
}

func RawJS(data string) *JS {
	return &JS{template.JS(data), true}
}

func NewJS(data string) *JS {
	return &JS{Data: template.JS(data)}
}
