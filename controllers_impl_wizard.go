package admin

// CRUD
type WizardController struct {
	CreateController
	UpdateController
	ReadController
}

func NewWizardController() *WizardController {
	return &WizardController{}
}

type ResourceWithCreateWizardController struct {
	ReadController
	UpdateController
	DeleteBulkController
	RestoreController
	IndexController
	SearchController
}
