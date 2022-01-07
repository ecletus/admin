package admin

type (
	WizardModelInterface interface {
		GetSteps() []string
		AddStep(name string)
		SetSteps(steps []string)
		CurrentStepName() string
		IsMainStep() bool
		Done()
		IsDone() bool
		IsGoingToBack() bool
		GetDestinationID() string
		SetDestinationID(id string)
		GetDestination() interface{}
		SetDestination(dest interface{})
	}

	WizardModelStepsGetter interface {
		WzGetStepsNames() (steps []string, mainStep string)
	}

	WizardModelUpdater interface {
		WizardModelInterface
	}

	WizardcallbackCompleter interface {
		OnWizardComplete(ctx *WizardCompleteContext, saveToDest func(*WizardCompleteContext) error) (err error)
	}

	StepSaver interface {
		WzStepCanSave() bool
	}

	StepLoadHandler interface {
		WzStepLoadHandle(ctx *WizardContext) (result StepLoadHandleResult, err error)
	}

	Steper interface {
		Next(ctx *WizardContext) (nextStep *NextStep, err error)
		SaveToDestination(ctx *WizardContext, destination interface{}) (err error)
	}

	StepDestinationCleaner interface {
		Steper
		CleanDestination(ctx *WizardContext, destination interface{})
	}

	StepRecordReader interface {
		Steper
		ReadRecord(ctx *WizardContext, record interface{}) (err error)
	}

	StepGobackAcceptor interface {
		AcceptGoback(ctx *WizardContext) bool
	}
)
