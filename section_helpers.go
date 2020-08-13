package admin

import (
	"strings"

	tag_scanner "github.com/unapu-go/tag-scanner"
)

func ParseSections(s string) (secs []*Section) {
	if s = strings.TrimSpace(s); s == "" {
		return
	}
	scanner := tag_scanner.Default
	if scanner.IsTags(s) {
		s = strings.ReplaceAll(s, "\n", ";")
	} else {
		s = "{" + strings.ReplaceAll(strings.ReplaceAll(s, "\n", "};{"), ",", ";") + "}"
	}
	secs = []*Section{{}}
	var (
		last = secs[0]
	)
	scanner.ScanAll(s, func(n tag_scanner.Node) {
		switch n.Type() {
		case tag_scanner.Flag:
			// new row
			// parse rows and columns
			s := n.(tag_scanner.NodeFlag).String()
			if scanner.IsTags(s) {
				// is a row
				var row []string
				scanner.ScanAll(s, func(n tag_scanner.Node) {
					if n.Type() == tag_scanner.Flag {
						row = append(row, n.String())
					}
				})
				if len(row) > 0 {
					if last.Title == "" {
						last.Rows = append(last.Rows, row)
					} else {
						last = &Section{Rows: [][]string{row}}
						secs = append(secs, last)
					}
				}
			} else {
				// only column
				if last.Title == "" {
					last.Rows = append(last.Rows, []string{s})
				} else {
					last = &Section{Rows: [][]string{{s}}}
					secs = append(secs, last)
				}
			}
		case tag_scanner.Tags:
			// new row
			// parse rows and columns
			var row []string
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
			last = &Section{Title: kv.Key}
			if scanner.IsTags(kv.Value) {
				// parse rows and columns
				scanner.ScanAll(kv.Value, func(n tag_scanner.Node) {
					switch n.Type() {
					case tag_scanner.Flag:
						// simple column
						last.Rows = append(last.Rows, []string{n.String()})
					case tag_scanner.Tags:
						// new row
						// parse rows and columns
						var row []string
						scanner.ScanAll(n.(tag_scanner.NodeTags).String(), func(n tag_scanner.Node) {
							if n.Type() == tag_scanner.Flag {
								row = append(row, n.String())
							}
						})
						if len(row) > 0 {
							last.Rows = append(last.Rows, row)
						}
					}
				})
			}

			if len(last.Rows) > 0 {
				secs = append(secs, last)
			}
		}
	})

	if len(last.Rows) == 0 {
		return nil
	}
	return
}
