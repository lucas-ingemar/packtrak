package core

import (
	"github.com/lucas-ingemar/packtrak/internal/manifest"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/samber/lo"
)

func FilterIncomingObjects(pkgs []string, pmManifest manifest.PmManifest, mType manifest.ManifestObjectType) (filteredObjs []string, err error) {
	pkgs = lo.Uniq(pkgs)
	for _, arg := range pkgs {
		var objs []string
		pkgs, deps, err := manifest.Filter(pmManifest)
		if err != nil {
			return nil, err
		}
		if mType == manifest.TypeDependency {
			objs = deps
		} else {
			objs = pkgs
		}
		if lo.Contains(objs, arg) {
			shared.PtermWarning.Printfln("'%s' is already present in manifest", arg)
			continue
		}
		filteredObjs = append(filteredObjs, arg)
	}
	return filteredObjs, nil
}
