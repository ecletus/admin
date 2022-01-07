package admin

import "strings"

func (this *Context) ParseUrl(s string) string {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) == 2 {
		switch parts[0] {
		case "admin":
			return this.URL(parts[1])
		case "site":
			return this.Top().URL(parts[1])
		}
	}
	return s
}

func (this *Context) InternalUrl(s string) bool {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) == 2 {
		switch parts[0] {
		case "admin", "site":
			return true
		}
	}
	return false
}
