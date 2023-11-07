package packagemanagers

import (
	"fmt"

	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/spf13/viper"
)

var (
	PackageManagersRegistered = []shared.PackageManager{&Dnf{}, &Go{}}
	PackageManagers           = []shared.PackageManager{}
)

func InitPackageManagerConfig() {
	for _, pm := range PackageManagersRegistered {
		viper.SetDefault(keyName(pm, "enabled"), true)
	}
}

func InitPackageManagers() {
	for _, pm := range PackageManagersRegistered {
		if viper.GetBool(keyName(pm, "enabled")) {
			//FIXME: Here we should also make the init checks
			if err := pm.InitCheckCmd(); err != nil {
				shared.PtermWarning.Printfln("Disabling %s manager: %s", pm.Name(), err.Error())
				viper.Set(keyName(pm, "enabled"), false)
				continue
			}
			PackageManagers = append(PackageManagers, pm)
		}
	}
}

func keyName(pm shared.PackageManager, key string) string {
	return fmt.Sprintf("managers.%s.%s", pm.Name(), key)
}
