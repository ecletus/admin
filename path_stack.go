package admin

import (
	"path"
	"strings"
)

type PathStack []string

func (this *PathStack) Add(pth string) func() {
	*this = append(*this, pth)
	return func() {
		*this = (*this)[0 : len(*this)-1]
	}
}

func (this PathStack) Abs(pth string) string {
	if len(this) == 0 || !strings.HasPrefix(pth, "./") {
		return pth
	}
	pth = path.Clean(path.Dir(this[len(this)-1]) + "/" + pth[2:])
	if strings.HasPrefix(pth, "..") {
		return ""
	}
	return pth
}

func (this PathStack) String() string {
	if len(this) == 0 {
		return ""
	}
	return "`" + strings.Join(this, "`/`") + "`"
}

func (this PathStack) StringMessage(msg string) string {
	if len(this) == 0 {
		return msg
	}
	return msg + " "+this.String()
}
