package applicationpackage

import "strings"

type PackageLevel struct{}

func NewPackageLevel() *PackageLevel {
	return &PackageLevel{}
}

func (pl *PackageLevel) Style(name string) string {
	if name == "" {
		return ""
	}
	return "(" + strings.ToUpper(name[:1]) + name[1:] + ") "
}
