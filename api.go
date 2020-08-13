package admin

type ContextStringer interface {
	String(ctx *Context) string
}