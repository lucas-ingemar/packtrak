package machinery

import (
	"testing"

	"github.com/lucas-ingemar/packtrak/internal/packagemanagers"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/stretchr/testify/assert"
)

func TestTotalUpdatedDeps(t *testing.T) {
	// name := "Gladys"
	// if !want.MatchString(msg) || err != nil {
	//     t.Fatalf(`Hello("Gladys") = %q, %v, want match for %#q, nil`, msg, err, want)
	// }
	ds := map[string]shared.DependenciesStatus{}
	ds["go"] = shared.DependenciesStatus{
		Synced:  []shared.Dependency{{Name: "test0", FullName: "test0_full"}},
		Updated: []shared.Dependency{{Name: "test1", FullName: "test1_full"}},
		Missing: []shared.Dependency{{Name: "test2", FullName: "test2_full"}},
		Removed: []shared.Dependency{{Name: "test3", FullName: "test3_full"}},
	}
	tud := TotalUpdatedDeps([]shared.PackageManager{&packagemanagers.Go{}}, ds)
	assert.Len(t, tud, 3, "number of dependencies")
}
