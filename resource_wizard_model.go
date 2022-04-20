package admin

import (
	"fmt"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/go-aorm/aorm"
	"github.com/go-aorm/aorm/types"
)

type WizardModel struct {
	aorm.Model
	aorm.Audited
	Steps         types.Strings
	isDone        bool
	GoBack        bool   `aorm:"-"`
	DestinationID string `admin:"-"`
	destination   interface{}
}

func (this *WizardModel) GetDestination() interface{} {
	return this.destination
}

func (this *WizardModel) SetDestination(dest interface{}) {
	this.destination = dest
}

func (this *WizardModel) SetDestinationID(id string) {
	this.DestinationID = id
}

func (this *WizardModel) GetDestinationID() string {
	return this.DestinationID
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
			mv := metaValues.ByName["GoBack"]
			metaValues.Values = []*resource.MetaValue{mv}
			metaValues.ByName = map[string]*resource.MetaValue{"GoBack": mv}
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

func (this *WizardModel) IsMainStep() bool {
	return len(this.Steps) == 1
}

func (this *WizardModel) IsGoingToBack() bool {
	return this.GoBack
}
