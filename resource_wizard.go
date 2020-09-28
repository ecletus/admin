package admin

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/moisespsena-go/bid"
	"github.com/pkg/errors"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"

	"github.com/moisespsena-go/aorm"
	"github.com/moisespsena-go/aorm/types"
)

func (this *Resource) AddCreateWizard(value interface{}, config ...*Config) *Resource {
	var cfg *Config
	for _, cfg = range config {
	}

	if cfg == nil {
		cfg = &Config{}
	}

	cfg.Param = this.Param + "/new"
	cfg.Wizard = &Wizard{}

	if cfg.Permission == nil {
		cfg.Permission = this.Config.Permission
	}

	if cfg.ViewControllerFactory == nil {
		cfg.ViewControllerFactory = func(controller interface{}) interface{} {
			return NewMainController(nil, controller)
		}
	}

	res := this.Admin.AddResource(value, cfg)
	this.SetCreateWizard(res)
	return res
}

func (this *Resource) SetCreateWizard(res *Resource) {
	res.Config.Wizard.BaseResource = this
	res.Config.Wizard.Resource = res
	res.Config.Wizard.Configure()
	this.createWizard = res.Config.Wizard
	res.I18nPrefix = this.I18nPrefix
}

func (this *Resource) CreateWizard() *Wizard {
	return this.createWizard
}

type Wizard struct {
	BaseResource, Resource *Resource
	GetStepsFunc           func() (steps []string, mainStep string)
	steps                  []string
	mainStep               string
	DefaultEditSections    func(ctx *Context, record interface{}) []*Section
}

func (this *Wizard) MainStep() string {
	return this.mainStep
}

func (this *Wizard) Configure() {
	this.steps = nil

	if this.GetStepsFunc != nil {
		this.steps, this.mainStep = this.GetStepsFunc()
	} else if stepsGetter, ok := this.Resource.ModelStruct.Value.(WizardModelStepsGetter); ok {
		this.steps, this.mainStep = stepsGetter.WzGetStepsNames()
	} else {
		for _, child := range this.Resource.ModelStruct.Children {
			meta := this.Resource.Meta(&Meta{Name: child.ParentField.Name})
			if sv := meta.Tags.Get("STEP"); sv == "STEP" {
				this.steps = append(this.steps, meta.Name)
			} else if sv == "main" {
				if this.mainStep != "" {
					panic("duplicate main step")
				}
				this.mainStep = meta.Name
				this.steps = append(this.steps, meta.Name)
			}
		}
	}

	if this.mainStep == "" {
		panic("no main step defined")
	}

	if len(this.steps) == 0 {
		panic("no steps defined")
	}

	for _, step := range this.steps {
		this.Resource.MetaRequired(step)
	}

	this.Resource.NewSectionsFunc = func(ctx *Context) []*Section {
		return []*Section{{Resource: this.Resource, Rows: [][]string{{this.mainStep}}}}
	}
	this.Resource.EditSectionsFunc = func(ctx *Context, record interface{}) (sections []*Section) {
		wz := record.(WizardModelInterface)
		if this.DefaultEditSections != nil {
			sections = this.DefaultEditSections(ctx, record)
		}
		sections = append(sections, []*Section{{Resource: this.Resource, Rows: [][]string{{wz.CurrentStepName()}}}}...)
		return
	}
	this.Resource.GetContextAttrsFunc = func(ctx *Context) []string {
		if ctx.Type == NEW {
			return []string{this.mainStep}
		} else if ctx.Type == EDIT {
			wz := ctx.Result.(WizardModelInterface)
			return []string{wz.CurrentStepName(), "GoBack"}
		}
		return nil
	}

	this.Resource.Meta(&Meta{Name: "GoBack"}).DisableSiblingsRequirement = resource.SiblingsRequirementCheckDisabledOnFalse

	this.Resource.OnBeforeCreate(func(ctx *core.Context, record interface{}) error {
		wz := record.(WizardModelInterface)
		wz.AddStep(this.mainStep)
		currentStepValue := reflect.ValueOf(record).Elem().FieldByName(this.mainStep)
		currentStepRecord := currentStepValue.Interface()

		wzContext := &WizardContext{
			Context:  ContextFromContext(ctx),
			StepName: this.mainStep,
			Wizard:   this,
			WzRecord: wz,
		}

		nextStep, err := currentStepRecord.(Steper).Next(wzContext)
		if err != nil {
			return err
		}
		wz.AddStep(nextStep.Name)
		return nil
	})

	this.Resource.OnAfterUpdate(func(ctx *core.Context, old, record interface{}) (err error) {
		wz := record.(WizardModelInterface)
		if wz.IsDone() {
			return this.SaveToRecord(ContextFromContext(ctx), wz)
		}
		return
	})

	this.Resource.OnBeforeUpdate(func(ctx *core.Context, _, record interface{}) error {
		wz := record.(WizardModelInterface)
		return this.Save(ContextFromContext(ctx), wz)
	})

	this.Resource.OnAfterFindOne(func(ctx *core.Context, record interface{}) error {
		return this.LoadHandle(ContextFromContext(ctx), record.(WizardModelInterface))
	})
}

type StepField struct {
	Model      WizardModelInterface
	ModelValue reflect.Value
	Name       string
	StepValue  reflect.Value
	Step       Steper
}

func (this *StepField) SetZero() {
	aorm.SetZero(this.StepValue)
	this.Step = nil
}

func (this *Wizard) Steps(wz WizardModelInterface) (fields []*StepField) {
	steps := wz.GetSteps()
	fields = make([]*StepField, len(steps))
	recordValue := reflect.ValueOf(wz)
	if recordValue.Kind() == reflect.Interface {
		recordValue = recordValue.Elem()
	}

	for i, name := range steps {
		currentStepValue := recordValue.Elem().FieldByName(name)
		currentStepRecord := currentStepValue.Interface().(Steper)
		if currentStepValue.IsNil() {
			fields[i] = &StepField{
				Model:      wz,
				ModelValue: recordValue,
				Name:       name,
			}
			continue
		}
		f := &StepField{
			Model:      wz,
			ModelValue: recordValue,
			Name:       name,
			StepValue:  currentStepValue,
			Step:       currentStepRecord,
		}
		fields[i] = f
	}
	return
}

func (this *Wizard) LoadHandle(ctx *Context, wz WizardModelInterface) (err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, "Wizard: load handle")
		}
	}()
	wzContext := &WizardContext{
		Context:  ContextFromContext(ctx),
		Wizard:   this,
		WzRecord: wz,
	}

	var (
		newSteps   types.Strings
		stepFields = this.Steps(wz)
	)

	for i, field := range stepFields {
		if field.Step == nil {
			newSteps = append(newSteps, field.Name)
			continue
		}

		if handler, ok := field.Step.(StepLoadHandler); ok {
			wzContext.StepName = field.Name
			var result StepLoadHandleResult
			if result, err = handler.WzStepLoadHandle(wzContext); err != nil {
				return errors.Wrapf(err, "step %q", field.Name)
			}
			if result.Skip {
				if field.Name == this.mainStep {
					return errors.New("the main step can be not skipped.")
				}
				var prev = stepFields[i-1]
				var nextResult *NextStep
				if nextResult, err = prev.Step.Next(wzContext); err != nil {
					return errors.Wrapf(err, "Step %s", prev.Name)
				}
				if nextResult == nil {
					// remove este e todos os steps posteriores
					for i := i; i < len(stepFields); i++ {
						stepFields[i].SetZero()
						stepFields[i] = nil
					}
					break
				}
				if nextResult.Name == field.Name {
					// este step NÃO será removido, nada muda
					newSteps = append(newSteps, field.Name)
				} else if i == len(stepFields)-1 {
					// não há steps posteriores, apenas este será removido e o novo step será adicionado
					newSteps = append(newSteps, nextResult.Name)
					if nextResult.Value != nil {
						field.ModelValue.Elem().FieldByName(nextResult.Name).Set(reflect.ValueOf(nextResult.Value))
					}
					stepFields[i].SetZero()
					break
				} else {
					next := stepFields[i+1]
					if nextResult.Name == next.Name {
						// havendo posteriores, para `PrevStep.Next(...).Index == NextStep.Index`, apenas remove este step
						stepFields[i].SetZero()
						continue
					}
					// havendo posteriores, para `PrevStep.Next(...).Index != NextStep.Index`, remove este e os posteriores,
					// e adiciona o novo step
					newSteps = append(newSteps, nextResult.Name)
					if nextResult.Value != nil {
						field.ModelValue.Elem().FieldByName(nextResult.Name).Set(reflect.ValueOf(nextResult.Value))
					}
					// remove este e todos os steps posteriores
					for i := i; i < len(stepFields); i++ {
						stepFields[i].SetZero()
						stepFields[i] = nil
					}
				}
			} else {
				newSteps = append(newSteps, field.Name)
			}
		} else {
			newSteps = append(newSteps, field.Name)
		}
	}

	if len(newSteps) != len(stepFields) {
		wz.SetSteps(newSteps)
	}
	return err
}

func (this *Wizard) Save(ctx *Context, wz WizardModelInterface) (err error) {
	recordValue := reflect.ValueOf(wz)
	if wz.IsGoingToBack() {
		steps := wz.GetSteps()
		wzContext := &WizardContext{
			Context:  ContextFromContext(ctx),
			Wizard:   this,
			WzRecord: wz,
		}
		for i := len(steps) - 1; i >= 0; i-- {
			currentStepValue := recordValue.Elem().FieldByName("Step_" + steps[i])
			currentStepRecord := currentStepValue.Interface()
			if acceptor, ok := currentStepRecord.(StepGobackAcceptor); ok {
				if !acceptor.AcceptGoback(wzContext) {
					continue
				}
			}
			steps = steps[0:i]
			break
		}
		wz.SetSteps(steps)
	} else {
		currentStepName := wz.CurrentStepName()
		currentStepValue := recordValue.Elem().FieldByName(currentStepName)
		currentStepRecord := currentStepValue.Interface()

		wzContext := &WizardContext{
			Context:  ContextFromContext(ctx),
			StepName: currentStepName,
			Wizard:   this,
			WzRecord: wz,
		}

		step := currentStepRecord.(Steper)
		nextStep, err := step.Next(wzContext)

		if err != nil {
			return err
		} else if nextStep == nil {
			wz.Done()
		} else {
			if saver, ok := step.(StepSaver); ok && !saver.WzStepCanSave() {
				steps := wz.GetSteps()
				wz.SetSteps(steps[0 : len(steps)-1])

			}
			wz.AddStep(nextStep.Name)
			if nextStep.Value != nil {
				recordValue.Elem().FieldByName(nextStep.Name).Set(reflect.ValueOf(nextStep.Value))
			} else {
				nextStepField := recordValue.Elem().FieldByName(nextStep.Name)
				if nextStepField.IsNil() {
					nextStepField.Set(reflect.New(nextStepField.Type().Elem()))
				}
			}
		}
	}
	steps := wz.GetSteps()
stepsLoop:
	for _, name := range this.steps {
		for _, step := range steps {
			if step == name {
				continue stepsLoop
			}
		}
		f := recordValue.Elem().FieldByName(name)
		f.Set(reflect.Zero(f.Type()))
	}
	return nil
}

func (this *Wizard) SaveToRecord(ctx *Context, wz WizardModelInterface) (err error) {
	dest := this.BaseResource.NewStruct(ctx.Site)
	destDB := ctx.DB().ModelStruct(this.BaseResource.ModelStruct)
	recordValue := reflect.ValueOf(wz)
	if recordValue.Kind() == reflect.Interface {
		recordValue = recordValue.Elem()
	}

	wzContext := &WizardContext{
		Context:  ctx,
		WzRecord: wz,
	}
	stepRecord := func(stepName string, create bool) Steper {
		currentStepValue := recordValue.Elem().FieldByName(stepName)
		if create {
			currentStepValue.Set(reflect.New(currentStepValue.Type().Elem()))
		}
		wzContext.StepName = stepName
		return currentStepValue.Interface().(Steper)
	}
	addStep := func(stepIndex string) (err error) {
		step := stepRecord(stepIndex, false)
		if err = step.SaveToDestination(wzContext, dest); err != nil {
			return errors.Wrapf(err, "step %d: add to record", stepIndex)
		}
		return
	}
	if destIdValuers := wz.GetDestinationID(); len(destIdValuers) > 0 {
		if err = aorm.SetIDValuersToRecord(this.BaseResource.ModelStruct, dest, destIdValuers); err != nil {
			return errors.Wrapf(err, "wizard: load destination: set destination id")
		}

		if err = destDB.First(dest).Error; err != nil {
			return errors.Wrapf(err, "wizard: load destination")
		}
		wzContext.Updates = true
		steps := wz.GetSteps()

	steps:
		for ix, name := range this.steps {
			for _, step := range steps {
				if step == strconv.Itoa(ix) {
					// add or update
					if err = addStep(step); err != nil {
						return
					}
					continue steps
				}
			}
			// remove old
			f := recordValue.Elem().FieldByName(name)
			f.Set(reflect.Zero(f.Type()))
			step := stepRecord(strconv.Itoa(ix), true)
			if remover, ok := step.(StepDestinationCleaner); ok {
				remover.CleanDestination(wzContext, dest)
			}
		}
	} else {
		for _, step := range wz.GetSteps() {
			if err = addStep(step); err != nil {
				return
			}
		}
	}
	if err = destDB.Save(dest).Error; err == nil {
		wz.SetDestinationID(this.BaseResource.ModelStruct.GetID(dest).Values())
		wz.SetDestination(dest)
		// err = ctx.DB().ModelStruct(this.Resource.ModelStruct).Delete(record).Error
	}
	return
}

func (this *Wizard) RemoveFromRecord(ctx *Context) {
	// 1: passa por cada etapa nao valida e executa o Remove From Record
}

type WizardContext struct {
	*Context
	StepName string
	Path     []string
	Updates  bool
	Wizard   *Wizard
	WzRecord WizardModelInterface
}

type WizardStepModel struct {
	aorm.Model
	aorm.Timestamps
}

type WizardModelInterface interface {
	GetSteps() []string
	AddStep(name string)
	SetSteps(steps []string)
	CurrentStepName() string
	Done()
	IsDone() bool
	IsGoingToBack() bool
	GetDestinationID() []aorm.IDValuer
	SetDestinationID(id []aorm.IDValuer)
	GetDestination() interface{}
	SetDestination(dest interface{})
}

type WizardModelStepsGetter interface {
	WzGetStepsNames() (steps []string, mainStep string)
}

type WizardModelUpdater interface {
	WizardModelInterface
}

type WizardModel struct {
	aorm.Model
	aorm.Audited
	Steps         types.Strings
	isDone        bool
	GoBack        bool    `aorm:"-"`
	DestinationID bid.BID `admin:"-"`
	destination   interface{}
}

func (this *WizardModel) GetDestination() interface{} {
	return this.destination
}

func (this *WizardModel) SetDestination(dest interface{}) {
	this.destination = dest
}

func (this *WizardModel) SetDestinationID(id []aorm.IDValuer) {
	if len(id) == 0 {
		this.DestinationID = nil
	} else {
		this.DestinationID = id[0].Raw().(bid.BID)
	}
}

func (this *WizardModel) GetDestinationID() []aorm.IDValuer {
	if !this.DestinationID.IsZero() {
		return []aorm.IDValuer{aorm.BIDIdValuer(this.DestinationID)}
	}
	return nil
}

func (this *WizardModel) AormAfterStructSetup(model *aorm.ModelStruct) {
	// remove all unique indexes
	var do func(model *aorm.ModelStruct)
	do = func(model *aorm.ModelStruct) {
		model.UniqueIndexes = nil
		for _, child := range model.Children {
			do(child)
		}
	}
	do(model)
}

func (this *WizardModel) BeforeCommitMetaValues(ctx *core.Context, res resource.Resourcer, metaValues *resource.MetaValues) {
	if metaValues.GetString("GoBack") == "true" {
		this.GoBack = true
		if len(metaValues.Values) > 1 {
			metaValues.Values = []*resource.MetaValue{metaValues.Values[metaValues.ByName["GoBack"]]}
			metaValues.ByName = map[string]int{"GoBack": 0}
		}
	}
}

func (this *WizardModel) Done() {
	this.isDone = true
}

func (this *WizardModel) IsDone() bool {
	return this.isDone
}

func (this *WizardModel) GetSteps() []string {
	return this.Steps
}

func (this *WizardModel) BackStep() {
	this.Steps = this.Steps[0 : len(this.Steps)-1]
}

func (this *WizardModel) SetSteps(steps []string) {
	var hasm = map[string]bool{}
	for _, s := range steps {
		if _, ok := hasm[s]; !ok {
			hasm[s] = true
		} else {
			panic(fmt.Errorf("recursive wizard step detected: %v", steps))
		}
	}
	this.Steps = steps
}

func (this *WizardModel) AddStep(name string) {
	this.SetSteps(append(this.Steps, name))
}
func (this *WizardModel) CurrentStepName() string {
	if len(this.Steps) == 0 {
		return ""
	}
	return this.Steps[len(this.Steps)-1]
}

func (this *WizardModel) IsGoingToBack() bool {
	return this.GoBack
}

type StepSaver interface {
	WzStepCanSave() bool
}

// StepLoadHandleResult resultado do StepLoadHandler.WzStepLoadHandle
type StepLoadHandleResult struct {
	// Skip pula este step.
	// Somente steps de indice >= 1 podem ser pulados.
	// O `Steper.Next` do Step anterior a este será invocado e:
	// - caso `PrevStep.Next(...) == nil`: remove este e todos os steps posteriores.
	// - caso `PrevStep.Next(...).Index == this.Index`: este step NÃO será removido.
	// - caso `PrevStep.Next(...).Index != this.Index`:
	//    - se não houver steps posteriores, apenas este será removido e o novo step será adicionado
	//    - havendo posteriores, para `PrevStep.Next(...).Index == NextStep.Index`, apenas remove este step
	//    - havendo posteriores, para `PrevStep.Next(...).Index != NextStep.Index`, remove este e os posteriores,
	//      e adiciona o novo step
	Skip bool
}

type StepLoadHandler interface {
	WzStepLoadHandle(ctx *WizardContext) (result StepLoadHandleResult, err error)
}

type Steper interface {
	Next(ctx *WizardContext) (nextStep *NextStep, err error)
	SaveToDestination(ctx *WizardContext, destination interface{}) (err error)
}

type StepDestinationCleaner interface {
	Steper
	CleanDestination(ctx *WizardContext, destination interface{})
}

type StepRecordReader interface {
	Steper
	ReadRecord(ctx *WizardContext, record interface{}) (err error)
}

type StepGobackAcceptor interface {
	AcceptGoback(ctx *WizardContext) bool
}

type NextStep struct {
	Name  string
	Value interface{}
}
