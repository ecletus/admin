package admin

// CRUD
type CrudController struct {
	CreateController
	ReadController
	UpdateController
	DeleteBulkController
	RestoreController
}

type CrudIndexController struct {
	CrudController
	IndexController
}

type CrudSearchController struct {
	CrudController
	SearchController
}

type CrudSearchIndexController struct {
	CrudSearchController
	IndexController
}

func NewCrudSearchIndexController() *CrudSearchIndexController {
	return &CrudSearchIndexController{}
}
