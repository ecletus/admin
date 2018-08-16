package admin

import (
	"github.com/moisespsena/go-i18n-modular/i18nmod"
	"github.com/aghape/helpers"
)

var (
	PKG       = helpers.GetCalledDir()
	I18NGROUP = i18nmod.PkgToGroup(PKG)
)
