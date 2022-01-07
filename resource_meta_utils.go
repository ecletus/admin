package admin

func SetMetaAsRequired(res *Resource, name ...string) {
	for _, name := range name {
		res.Meta(&Meta{Name: name}).SetRequired(true)
	}
}
