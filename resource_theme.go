package admin

// UseTheme use them for resource, will auto load the theme's javascripts, stylesheets for this resource
func (this *Resource) UseTheme(theme interface{}) []ThemeInterface {
	var themeInterface ThemeInterface
	if ti, ok := theme.(ThemeInterface); ok {
		themeInterface = ti
	} else if str, ok := theme.(string); ok {
		for _, theme := range this.Config.Themes {
			if theme.GetName() == str {
				return this.Config.Themes
			}
		}

		themeInterface = Theme{Name: str}
	}

	if themeInterface != nil {
		this.Config.Themes = append(this.Config.Themes, themeInterface)
		themeInterface.ConfigAdminTheme(this)
	}
	return this.Config.Themes
}

// GetTheme get registered theme with name
func (this *Resource) GetTheme(name string) ThemeInterface {
	for _, theme := range this.Config.Themes {
		if theme.GetName() == name {
			return theme
		}
	}
	return nil
}
