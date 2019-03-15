package admin

import "fmt"

// UseTheme use them for resource, will auto load the theme's javascripts, stylesheets for this resource
func (res *Resource) UseDisplay(display interface{}) {
	var displayInterface DisplayInterface
	if ti, ok := display.(DisplayInterface); ok {
		displayInterface = ti
	} else if str, ok := display.(string); ok {
		if res.GetDisplay(str) != nil {
			return
		}

		displayInterface = &Display{Name: str}
	}

	if displayInterface != nil {
		if res.Config.Displays == nil {
			res.Config.Displays = make(map[string]DisplayInterface)
		}
		res.Config.Displays[displayInterface.GetName()] = displayInterface
		displayInterface.ConfigAdminTheme(res)
	}
}

func (res *Resource) GetDefaultDisplayName() string {
	if res.defaultDisplayName == "" {
		return "default"
	}
	return res.defaultDisplayName
}

func (res *Resource) SetDefaultDisplay(displayName string) {
	display := res.GetDisplay(displayName)
	if display == nil {
		panic(fmt.Errorf("Display %q does not exists.", displayName))
	}
	res.defaultDisplayName = displayName
}

func (res *Resource) GetDefaultDisplay() DisplayInterface {
	display := res.GetDisplay(res.GetDefaultDisplayName())
	if display == nil {
		return DefaultDisplay
	}
	return display
}

// GetDisplay get registered theme with name
func (res *Resource) GetDisplay(name string) DisplayInterface {
	if res.Config.Displays != nil {
		if d, ok := res.Config.Displays[name]; ok {
			return d
		}
	}
	return nil
}
