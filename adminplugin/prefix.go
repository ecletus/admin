package adminplugin

import "github.com/aghape/helpers"

func prefix() string {
	return helpers.GetCalledDir()
}

var PREFIX = prefix()
