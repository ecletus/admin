package admin

var DefaultDisplay = &Display{
	Name:  "default",
	Label: PKG + ".resource.display.Default",
}

type DisplayInterface interface {
	ThemeInterface
	GetLayoutName() string
	GetIcon() string
	GetLabel() string
}

type Display struct {
	Name                 string
	LayoutName           string
	Icon                 string
	Label                string
	GetViewPathsFunc     func() []string
	ConfigAdminThemeFunc func(*Resource)
}

// GetName get name from theme
func (d *Display) GetName() string {
	return d.Name
}

// GetViewPaths get view paths from theme
func (d *Display) GetViewPaths() []string {
	if d.GetViewPathsFunc != nil {
		return d.GetViewPathsFunc()
	}
	return []string{}
}

// ConfigAdminTheme config theme for admin resource
func (d *Display) ConfigAdminTheme(res *Resource) {
	if d.ConfigAdminThemeFunc != nil {
		d.ConfigAdminThemeFunc(res)
	}
	return
}

func (d *Display) GetLayoutName() string {
	if d.LayoutName == "" {
		return DEFAULT_LAYOUT
	}
	return d.LayoutName
}

func (d *Display) GetIcon() string {
	return d.Icon
}

func (d *Display) GetLabel() string {
	return d.Label
}
