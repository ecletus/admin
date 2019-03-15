package admin

// UseTheme use them for resource, will auto load the theme's javascripts, stylesheets for this resource
func (res *Resource) UseTheme(theme interface{}) []ThemeInterface {
	var themeInterface ThemeInterface
	if ti, ok := theme.(ThemeInterface); ok {
		themeInterface = ti
	} else if str, ok := theme.(string); ok {
		for _, theme := range res.Config.Themes {
			if theme.GetName() == str {
				return res.Config.Themes
			}
		}

		themeInterface = Theme{Name: str}
	}

	if themeInterface != nil {
		res.Config.Themes = append(res.Config.Themes, themeInterface)
		themeInterface.ConfigAdminTheme(res)
	}
	return res.Config.Themes
}

// GetTheme get registered theme with name
func (res *Resource) GetTheme(name string) ThemeInterface {
	for _, theme := range res.Config.Themes {
		if theme.GetName() == name {
			return theme
		}
	}
	return nil
}
