package admin

import (
	"fmt"
	"sync"
)

type PageInterface interface {
	Path() string
	Title() string
	Description() string
	Serve(ctx *Context)
}

type Paged struct {
	pages map[string]PageInterface
	mu    sync.Mutex
}

func (p *Paged) AddPage(page ...PageInterface) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.pages == nil {
		p.pages = map[string]PageInterface{}
	}

	for _, page := range page {
		path := page.Path()
		if old, ok := p.pages[path]; ok {
			panic(fmt.Errorf("admin.Paged.AddPage: duplicate page %q: {old: %T, new: %T}", path, old, page))
		}
		p.pages[page.Path()] = page
	}
}

func (p *Paged) GetPageOk(path string) (page PageInterface, ok bool) {
	if p.pages == nil {
		return
	}
	page, ok = p.pages[path]
	return 
}

func (p *Paged) GetPage(path string) (page PageInterface) {
	if p.pages == nil {
		return nil
	}
	return p.pages[path]
}

func (p *Paged) HasPage(path string) (page PageInterface) {
	if p.pages == nil {
		return nil
	}
	return p.pages[path]
}