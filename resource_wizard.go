package admin

import (
	"reflect"
	"regexp"
	"strconv"

	"github.com/pkg/errors"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"

	"github.com/go-aorm/aorm"
	"github.com/go-aorm/aorm/types"
)

func (this *Resource) AddCreateWizard(value interface{}, config ...*Config) *Resource {
	var cfg *Config
	for _, cfg = range config {
	}

	if cfg == nil {
		cfg = &Config{}
	}

	var name string

	if len(this.createWizards) == 0 {
		cfg.Param = "new"
		name = "default"
	} else {
		if name = cfg.Param; name == "" {
			name = strconv.Itoa(len(this.createWizards))
		}
		cfg.Param = "new/" + name
	}
	cfg.Wizard = &Wizard{}

	if cfg.Permission == nil {
		cfg.Permission = this.Config.Permission
	}

	if cfg.ViewControllerFactory == nil {
		cfg.ViewControllerFactory = func(controller interface{}) interface{} {
			return NewMainController(nil, controller)
		}
	}

	res := this.AddResource(&SubConfig{Parent: this, MountAsItemDisabled: true}, value, cfg)
	res.Config.Wizard.Resource = res
	res.Config.Wizard.Configure()
	this.createWizards = append(this.createWizards, res.Config.Wizard)
	res.I18nPrefix = this.I18nPrefix

	if this.CreateWizardByName == nil {
		this.CreateWizardByName = map[string]*Wizard{}
	}
	this.CreateWizardByName[name] = res.Config.Wizard
	return res
}

func (this *Resource) NewWizardRecord(name string, rec ...WizardModelInterface) (model *aorm.ModelStruct, record WizardModelInterface) {
	wz := this.CreateWizardByName[name]
	return wz.Resource.ModelStruct, wz.New(rec...)
}

func (this *Resource) CreateWizards() []*Wizard {
	return this.createWizards
}

type Wizard struct {
	Resource            *Resource
	GetStepsFunc        func() (steps []string, mainStep string)
	steps               []string
	mainStep            string
	DefaultEditSections func(ctx *Context, record interface{}) []*Section
	CompleteCallback    func(ctx *WizardCompleteContext, saveToDest func(*WizardCompleteContext) error) (err error)
}

func (this *Wizard) New(rec ...WizardModelInterface) (record WizardModelInterface) {
	for _, record = range rec {
	}
	if record == nil {
		record = this.Resource.New().(WizardModelInterface)
	}
	record.SetSteps([]string{this.mainStep})
	return
}

func (this *Wizard) CreateSibling(name string, rec ...WizardModelInterface) (model *aorm.ModelStruct, record WizardModelInterface) {
	return this.Resource.ParentResource.NewWizardRecord(name, rec...)
}

func (this *Wizard) NewStep(wz WizardModelInterface, name string) Steper {
	return this.NewStepV(wz, reflect.ValueOf(wz).Elem(), name)
}

func (this *Wizard) NewStepV(wz WizardModelInterface, wzValue reflect.Value, name string) Steper {
	wz.AddStep(name)
	stepField := wzValue.FieldByName(name)
	if stepField.IsNil() {
		stepField.Set(reflect.New(stepField.Type().Elem()))
	}
	return stepField.Interface().(Steper)
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
			meta := this.Resource.Meta(&Meta{Name: child.ChildConfig.ParentField.Name})
			if meta.Tags.Flag("STEP") {
				this.steps = append(this.steps, meta.Name)
			} else if meta.Tags.GetString("STEP") == "main" {
				if this.mainStep != "" {
					panic("duplicate main step")
				}
				this.mainStep = meta.Name
				this.steps = append(this.steps, meta.Name)
			} else {
				continue
			}
			_ = child.Value.(Steper)
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

	this.Resource.Sections.Default.Screen.New.GetContextFunc = func(ctx *SectionsContext) []*Section {
		return []*Section{{Resource: this.Resource, Rows: [][]interface{}{{this.mainStep}}}}
	}
	this.Resource.Sections.Default.Screen.Edit.GetContextFunc = func(ctx *SectionsContext) (sections []*Section) {
		wz := ctx.Record.(WizardModelInterface)
		if this.DefaultEditSections != nil {
			sections = this.DefaultEditSections(ctx.Ctx, ctx.Record)
		}
		sections = append(sections, []*Section{{Resource: this.Resource, Rows: [][]interface{}{{wz.CurrentStepName()}}}}...)
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
			return this.SaveToDestination(ContextFromContext(ctx), wz)
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

type WizardContextCompleteConfig struct {
	RedirectTo           string
	RedirectToStatus     int
	RedirectToOther      bool
	FlashMessageDisabled bool
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
			currentStepValue := recordValue.Elem().FieldByName(steps[i])
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
			if nextStep.Value != nil {
				wz.AddStep(nextStep.Name)
				recordValue.Elem().FieldByName(nextStep.Name).Set(reflect.ValueOf(nextStep.Value))
			} else {
				this.NewStepV(wz, recordValue.Elem(), nextStep.Name)
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

func (this *Wizard) SaveToDestination(ctx *Context, wz WizardModelInterface) (err error) {
	ctx.WizardCompleteConfig = &WizardContextCompleteConfig{}

	wzCtx := &WizardCompleteContext{
		WizardContext{
			Context:  ctx,
			Wizard:   this,
			WzRecord: wz,
		},
		this.Resource.ParentResource.ModelStruct,
		this.Resource.ParentResource.NewStruct(ctx.Site),
	}
	if saver, ok := wzCtx.Dest.(WizardcallbackCompleter); ok {
		return saver.OnWizardComplete(wzCtx, this.saveToDestination)
	} else if this.CompleteCallback != nil {
		return this.CompleteCallback(wzCtx, this.saveToDestination)
	}
	return this.saveToDestination(wzCtx)
}

func (this *Wizard) saveToDestination(wzCtx *WizardCompleteContext) (err error) {
	ctx, wz, destModel, dest := wzCtx.Context, wzCtx.WzRecord, wzCtx.DestModel, wzCtx.Dest

	// loads all related many fields recursively to apply all values to destination
	if err = this.Resource.ModelStruct.LoadRelatedManyFields(ctx.DB(), wz); err != nil {
		return
	}

	destDB := ctx.DB().ModelStruct(destModel)
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
	if destId := wz.GetDestinationID(); destId != "" {
		destID, _ := destModel.ParseIDString(destId)
		destID.SetTo(dest)
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
		wz.SetDestinationID(destModel.GetID(dest).String())
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

type WizardCompleteContext struct {
	WizardContext
	DestModel *aorm.ModelStruct
	Dest      interface{}
}

func (this *WizardCompleteContext) RedirectToSibling(name string, record ...WizardModelInterface) (_ WizardModelInterface, err error) {
	model, rec := this.Wizard.CreateSibling(name, record...)
	aorm.MustCopyIdTo(this.Resource.ModelStruct.GetID(this.WzRecord), model.DefaultID()).SetTo(rec)

	if err = this.DB().ModelStruct(model).Save(rec).Error; err != nil {
		return
	}
	cfg := this.WizardCompleteConfig
	cfg.RedirectToOther = true
	cfg.RedirectTo = regexp.MustCompile("/(new|edit)/[^/?]+").ReplaceAllString(this.Request.RequestURI, "/$1/"+name+"/"+model.GetID(rec).String())
	cfg.FlashMessageDisabled = true
	return rec, nil
}

func (this *WizardCompleteContext) RedirectToIndex() {
	this.WizardCompleteConfig.RedirectTo = this.Resource.ParentResource.GetContextIndexURI(this.Context)
}

type WizardStepModel struct {
	aorm.Model
	aorm.Timestamps
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

type NextStep struct {
	Name  string
	Value interface{}
}
