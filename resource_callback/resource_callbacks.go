package resource_callback

import (
	"fmt"
	"strings"
	"github.com/moisespsena-go/topsort"
	"github.com/ecletus/admin"
)

const DUPLICATION_OVERRIDE = 0
const DUPLICATION_ABORT = 1
const DUPLICATION_SKIP = 2

// Callbacks type is a slice of standard middleware handlers with methods
// to compose middleware chains and interface{}'s.
type Callbacks []*Callback

type Callback struct {
	Name    string
	Handler func(chain *ChainHandler)
	Before  []string
	After   []string
}

type CallbacksStack struct {
	ByName          map[string]*Callback
	Items           Callbacks
	Anonymous       Callbacks
	acceptAnonymous bool
	compiled        bool
}

func NewCallbacksStack(acceptAnonymous ...bool) *CallbacksStack {
	if len(acceptAnonymous) == 0 {
		acceptAnonymous = []bool{true}
	}
	return &CallbacksStack{
		ByName:          make(map[string]*Callback),
		acceptAnonymous: acceptAnonymous[0],
	}
}

func (stack *CallbacksStack) Copy() *CallbacksStack {
	byName := make(map[string]*Callback)

	for key, md := range stack.ByName {
		byName[key] = md
	}

	anonymous := make(Callbacks, len(stack.Anonymous))
	copy(anonymous, stack.Anonymous)

	items := make(Callbacks, len(stack.Items))
	copy(items, stack.Items)

	return &CallbacksStack{
		ByName:          byName,
		Items:           items,
		Anonymous:       anonymous,
		acceptAnonymous: stack.acceptAnonymous,
	}
}

func (stack *CallbacksStack) Override(items Callbacks, option int) *CallbacksStack {
	return NewCallbacksStack(stack.acceptAnonymous).Add(items, option)
}

func (stack *CallbacksStack) Has(name ...string) bool {
	for _, n := range name {
		if _, ok := stack.ByName[n]; !ok {
			return false
		}
	}
	return true
}

func (stack *CallbacksStack) Add(items Callbacks, option int) *CallbacksStack {
	if stack.ByName == nil {
		stack.ByName = make(map[string]*Callback)
	}

	for i, md := range items {
		if md.Name == "" {
			if stack.acceptAnonymous {
				stack.Anonymous = append(stack.Anonymous, md)
			} else {
				panic(fmt.Errorf("Item %v Name is empty.", i))
			}
		} else {
			if stack.Has(md.Name) {
				switch option {
				case DUPLICATION_ABORT:
					panic(fmt.Errorf("%q has be registered.", md.Name))
				case DUPLICATION_SKIP:
					continue
				case DUPLICATION_OVERRIDE:
				default:
					panic(fmt.Errorf("Invalid option %v.", option))
				}
			}
			stack.ByName[md.Name] = md
		}
	}
	return stack
}

func (stack *CallbacksStack) Build() *CallbacksStack {
	notFound := make(map[string][]string)

	graph := topsort.NewGraph()

	for _, md := range stack.ByName {
		graph.AddNode(md.Name)
	}

	for _, md := range stack.ByName {
		for _, to := range md.Before {
			if stack.Has(to) {
				graph.AddEdge(md.Name, to)
			} else {
				if _, ok := notFound[md.Name]; !ok {
					notFound[md.Name] = make([]string, 1)
				}
				notFound[md.Name] = append(notFound[md.Name], to)
			}
		}
		for _, from := range md.After {
			if stack.Has(from) {
				graph.AddEdge(from, md.Name)
			} else {
				if _, ok := notFound[md.Name]; !ok {
					notFound[md.Name] = make([]string, 1)
				}
				notFound[md.Name] = append(notFound[md.Name], from)
			}
		}
	}

	if len(notFound) > 0 {
		var msgs []string
		for n, items := range notFound {
			msgs = append(msgs, fmt.Sprintf("Required by %q: %v.", n, strings.Join(items, ", ")))
		}
		panic(fmt.Errorf("Dependency error:\n - %v\n", strings.Join(msgs, "\n - ")))
	}

	names, err := graph.DepthFirst()

	if err != nil {
		panic(fmt.Errorf("Topological sorter error: %v", err))
	}

	stack.Items = make(Callbacks, 0)

	// named middlewares at begin
	for _, name := range names {
		stack.Items = append(stack.Items, stack.ByName[name])
	}

	// named middlewares at end
	for _, md := range stack.Anonymous {
		stack.Items = append(stack.Items, md)
	}

	return stack
}

func (stack *CallbacksStack) Run(res *admin.Resource) {
	if !stack.compiled {
		stack.Build()
		stack.compiled = true
	}
	stack.Items.Run(res)
}