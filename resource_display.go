package admin

import "fmt"

// UseTheme use them for resource, will auto load the theme's javascripts, stylesheets for this resource
func (this *Resource) UseDisplay(display interface{}) {
	var displayInterface DisplayInterface
	if ti, ok := display.(DisplayInterface); ok {
		displayInterface = ti
	} else if str, ok := display.(string); ok {
		if this.GetDisplay(str) != nil {
			return
		}

		displayInterface = &Display{Name: str}
	}

	if displayInterface != nil {
		if this.Config.Displays == nil {
			this.Config.Displays = make(map[string]DisplayInterface)
		}
		this.Config.Displays[displayInterface.GetName()] = displayInterface
		displayInterface.ConfigAdminTheme(this)
	}
}

func (this *Resource) GetDefaultDisplayName() string {
	if this.defaultDisplayName == "" {
		return "default"
	}
	return this.defaultDisplayName
}

func (this *Resource) SetDefaultDisplay(displayName string) {
	display := this.GetDisplay(displayName)
	if display == nil {
		panic(fmt.Errorf("Display %q does not exists.", displayName))
	}
	this.defaultDisplayName = displayName
}

func (this *Resource) GetDefaultDisplay() DisplayInterface {
	display := this.GetDisplay(this.GetDefaultDisplayName())
	if display == nil {
		return DefaultDisplay
	}
	return display
}

// GetDisplay get registered theme with name
func (this *Resource) GetDisplay(name string) DisplayInterface {
	if this.Config.Displays != nil {
		if d, ok := this.Config.Displays[name]; ok {
			return d
		}
	}
	return nil
}
