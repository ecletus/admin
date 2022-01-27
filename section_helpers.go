package admin

import (
	"strings"

	tag_scanner "github.com/unapu-go/tag-scanner"
)

type SectionsOpts struct {
	Parent  string
	Exclude []string
}

func ParseSections(s string) (secs []*Section, opts *SectionsOpts) {
	opts = &SectionsOpts{}
	var parseSections func(root bool, s string) (secs []*Section)
	parseSections = func(root bool, s string) (secs []*Section) {
		if s = strings.TrimSpace(s); s == "" {
			return
		}
		scanner := tag_scanner.Default
		if scanner.IsTags(s) {
			s = strings.ReplaceAll(s, "\n", ";")
		} else {
			s = "{" + strings.ReplaceAll(strings.ReplaceAll(s, "\n", "};{"), ",", ";") + "}"
		}

		last := &Section{}
		secs = []*Section{last}
		scanner.ScanAll(s, func(n tag_scanner.Node) {
			switch n.Type() {
			case tag_scanner.Flag:
				// new row
				// parse rows and columns
				s := n.(tag_scanner.NodeFlag).String()
				if scanner.IsTags(s) {
					// is a row
					var row []interface{}
					scanner.ScanAll(s, func(n tag_scanner.Node) {
						if n.Type() == tag_scanner.Flag {
							row = append(row, n.String())
						}
					})
					if len(row) > 0 {
						if last.Title == "" {
							last.Rows = append(last.Rows, row)
						} else {
							last = &Section{Rows: [][]interface{}{row}}
							secs = append(secs, last)
						}
					}
				} else {
					// only column
					if last.Title == "" {
						last.Rows = append(last.Rows, []interface{}{s})
					} else {
						last = &Section{Rows: [][]interface{}{{s}}}
						secs = append(secs, last)
					}
				}
			case tag_scanner.Tags:
				// new row
				// parse rows and columns
				var row []interface{}
				scanner.ScanAll(n.(tag_scanner.NodeTags).String(), func(n tag_scanner.Node) {
					if n.Type() == tag_scanner.Flag {
						row = append(row, n.String())
					}
				})
				if len(row) > 0 {
					last.Rows = append(last.Rows, row)
				}
			case tag_scanner.KeyValue:
				kv := n.(tag_scanner.NodeKeyValue)
				if root {
					if kv.Key == "_" {
						value := kv.Value
						tags, flags := tag_scanner.Parse(scanner, value)
						opts.Exclude = tags.GetTags("EXCLUDE", tag_scanner.FlagPreserveKeys).Flags()
						if len(flags) > 0 {
							opts.Parent = flags.Strings()[0]
						}
						return
					}
				}
				sec := &Section{Title: kv.Key}
				if scanner.IsTags(kv.Value) {
					// parse rows and columns
					scanner.ScanAll(kv.Value, func(n tag_scanner.Node) {
						switch t := n.(type) {
						case tag_scanner.NodeFlag:
							// simple column
							sec.Rows = append(sec.Rows, []interface{}{t.String()})
						case tag_scanner.NodeTags:
							// new row
							// parse rows and columns
							var row []interface{}
							scanner.ScanAll(t.String(), func(n tag_scanner.Node) {
								if n.Type() == tag_scanner.Flag {
									row = append(row, n.String())
								}
							})
							if len(row) > 0 {
								sec.Rows = append(sec.Rows, row)
							}
						case tag_scanner.NodeKeyValue:
							secs := parseSections(false, t.Value)
							if len(secs) > 0 {
								secs[0].Title = t.Key
								sec.Rows = append(sec.Rows, []interface{}{secs[0]})
							}
						}
					})
				}

				if len(sec.Rows) > 0 {
					secs = append(secs, sec)
					last = &Section{}
					secs = append(secs, last)
				}
			}
		})

		var newSecs []*Section
		for _, sec := range secs {
			if len(sec.Rows) > 0 {
				newSecs = append(newSecs, sec)
			}
		}

		return newSecs
	}

	secs = parseSections(true, s)
	if opts.Parent == "" && len(opts.Exclude) == 0 {
		opts = nil
	}
	return
}
