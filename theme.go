package admin

// ThemeInterface theme interface
type ThemeInterface interface {
	GetName() string
	ConfigAdminTheme(*Resource)
	Enabled(ctx *Context) bool
}

// Theme base theme config struct
type Theme struct {
	Name        string
	EnabledFunc func(ctx *Context) bool
}

// GetName get name from theme
func (theme Theme) GetName() string {
	return theme.Name
}

// ConfigAdminTheme config theme for admin resource
func (Theme) ConfigAdminTheme(*Resource) {
	return
}

func (t Theme) Enabled(ctx *Context) bool {
	if t.EnabledFunc != nil {
		return t.EnabledFunc(ctx)
	}
	return true
}
