package admin

type ResourceCallback struct {
	Name string
	After []string
	Before []string
	Callback func(res *Resource)
}

type ResourceCallbacks struct {

}
