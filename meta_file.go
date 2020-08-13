package admin

import (
	"github.com/dustin/go-humanize"
	"github.com/ecletus/core/resource"
	"sort"
	"strings"
)

type FileConfig struct {
	Multiple bool
	Capture  bool
	Audio    bool
	Video    bool
	Image    bool
	Accept   []string
	accept   string
	MaxSize uint
}

func (this *FileConfig) AcceptAttribute() string {
	return this.accept
}

func (this *FileConfig) MaxSizeString() string {
	if this.MaxSize > 0 {
		return humanize.Bytes(uint64(this.MaxSize))
	}
	return ""
}

// ConfigureQorMeta configure select one meta
func (this *FileConfig) ConfigureQorMeta(metaor resource.Metaor) {
	meta := metaor.(*Meta)
	meta.Type = "file"
	var (
		acceptsm = map[string]interface{}{}
		accepts []string
	)
	for _, accept := range this.Accept {
		if _, ok := acceptsm[accept]; !ok {
			acceptsm[accept] = nil
		}
	}
	sort.Strings(accepts)
	this.accept = strings.Join(accepts, ",")
}
