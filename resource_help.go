package admin

func (this *Resource) GetHelpPair() (key string, defaul string) {
	key = this.HelpKey

	if key == "" {
		key = this.I18nPrefix + ".help~s"
		defaul = this.PluralHelp
	}

	return key, defaul
}

func (this *Resource) GetPluralHelpPair() (key string, defaul string) {
	key = this.PluralHelpKey

	if key == "" {
		key = this.I18nPrefix + ".help~p"
		defaul = this.PluralHelp
	}

	return key, defaul
}
