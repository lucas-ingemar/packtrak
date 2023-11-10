package core

import (
	"testing"

	"github.com/lucas-ingemar/packtrak/internal/manifest"
	"github.com/lucas-ingemar/packtrak/internal/packagemanagers"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/stretchr/testify/assert"
)

func TestFilterIncomingObjects(t *testing.T) {
	pkgs := []string{"pkg1", "pkg2", "pkg2", "pkg4", "pkg4", "pkg5"}
	pmManifest := manifest.PmManifest{
		Global: manifest.Global{
			Dependencies: []string{"dep1", "dep2", "dep3"},
			Packages:     []string{"pkg1", "pkg2", "pkg3"},
		},
		Conditional: []manifest.Conditional{},
	}
	filteredObjs, err := FilterIncomingObjects(pkgs, pmManifest, manifest.TypePackage)
	assert.Nil(t, err, "should be no error should")
	assert.Equal(t, []string{"pkg4", "pkg5"}, filteredObjs, "incorrect filtering")
}

func TestTotalUpdatedDeps(t *testing.T) {
	ds := map[string]shared.DependenciesStatus{}
	ds["go"] = shared.DependenciesStatus{
		Synced:  []shared.Dependency{{Name: "test0", FullName: "test0_full"}},
		Updated: []shared.Dependency{{Name: "test1", FullName: "test1_full"}},
		Missing: []shared.Dependency{{Name: "test2", FullName: "test2_full"}},
		Removed: []shared.Dependency{{Name: "test3", FullName: "test3_full"}},
	}
	tud := CountUpdatedDeps([]shared.PackageManager{&packagemanagers.Go{}}, ds)
	assert.Equal(t, 3, tud, "number of dependencies")
}
