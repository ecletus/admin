package admin

import "github.com/ecletus/core"

func MetaConfigBooleanSelect() MetaConfigInterface {
	return &SelectOneConfig{
		AllowBlank: true,
		Collection: func(ctx *core.Context) [][]string {
			p := I18NGROUP + ".form.bool."
			return [][]string{
				{"true", ctx.Ts(p+"true", "Yes")},
				{"false", ctx.Ts(p+"false", "No")},
			}
		},
	}
}
