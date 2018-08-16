package admin

// ThemeInterface theme interface
type ThemeInterface interface {
	GetName() string
	ConfigAdminTheme(*Resource)
}

// Theme base theme config struct
type Theme struct {
	Name string
}

// GetName get name from theme
func (theme Theme) GetName() string {
	return theme.Name
}

// ConfigAdminTheme config theme for admin resource
func (Theme) ConfigAdminTheme(*Resource) {
	return
}
