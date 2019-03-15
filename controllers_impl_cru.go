package admin

// CRU - Not Deletable
type CruController struct {
	CreateController
	ReadController
	UpdateController
}

type CruIndexController struct {
	CruController
	IndexController
}

type CruSearchController struct {
	CruController
	SearchController
}

type CruSearchIndexController struct {
	CruSearchController
	IndexController
}

func NewCruSearchIndexController() *CruSearchIndexController {
	return &CruSearchIndexController{}
}
