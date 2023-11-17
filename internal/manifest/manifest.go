package manifest

import (
	"bytes"
	"fmt"
	"os"

	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/samber/lo"

	"gopkg.in/yaml.v3"
)

type (
	ManifestConditionalType string
	ManifestObjectType      string
)

var (
	MConditionHost  ManifestConditionalType = "host"
	MConditionGroup ManifestConditionalType = "group"
	TypePackage     ManifestObjectType      = "package"
	TypeDependency  ManifestObjectType      = "dependency"
)

type ManifestFace interface {
	Save(filename string) error
	Pm(name shared.ManagerName) PmManifest
	AddConditional(oType ManifestObjectType, pmName shared.ManagerName, cType ManifestConditionalType, cValue string, objects []string) error
	RemoveConditional(oType ManifestObjectType, pmName shared.ManagerName, cType ManifestConditionalType, cValue string, objects []string) error
	AddGlobal(oType ManifestObjectType, pmName shared.ManagerName, objects []string) error
	RemoveGlobal(oType ManifestObjectType, pmName shared.ManagerName, objects []string) error

	AddToHost(toAdd []string, pmName shared.ManagerName, oType ManifestObjectType) error
	AddToGroup(toAdd []string, group string, pmName shared.ManagerName, oType ManifestObjectType) error
}

type Manifest struct {
	Dnf     PmManifest `yaml:"dnf"`
	Git     PmManifest `yaml:"git"`
	Go      PmManifest `yaml:"go"`
	Version string     `yaml:"_version"`
}

func (m *Manifest) Pm(name shared.ManagerName) PmManifest {
	return *m.pmPnt(name)
}

func (m *Manifest) AddConditional(oType ManifestObjectType, pmName shared.ManagerName, cType ManifestConditionalType, cValue string, objects []string) error {
	c, err := m.getOrAddConditional(pmName, cType, cValue)
	if err != nil {
		return err
	}
	switch oType {
	case TypePackage:
		c.Packages = append(c.Packages, objects...)
	case TypeDependency:
		c.Dependencies = append(c.Dependencies, objects...)
	}
	return nil
}

func (m *Manifest) RemoveConditional(oType ManifestObjectType, pmName shared.ManagerName, cType ManifestConditionalType, cValue string, objects []string) error {
	c, err := m.getOrAddConditional(pmName, cType, cValue)
	if err != nil {
		return err
	}

	switch oType {
	case TypePackage:
		c.Packages = lo.Filter(c.Packages, func(item string, index int) bool {
			return !lo.Contains(objects, item)
		})
	case TypeDependency:
		c.Dependencies = lo.Filter(c.Dependencies, func(item string, index int) bool {
			return !lo.Contains(objects, item)
		})
	}
	return nil
}

func (m *Manifest) AddGlobal(oType ManifestObjectType, pmName shared.ManagerName, objects []string) error {
	fmt.Println(oType, "manifest")
	pm := m.pmPnt(pmName)
	switch oType {
	case TypePackage:
		pm.Global.Packages = append(pm.Global.Packages, objects...)
	case TypeDependency:
		pm.Global.Dependencies = append(pm.Global.Dependencies, objects...)
	}
	return nil
}

func (m *Manifest) RemoveGlobal(oType ManifestObjectType, pmName shared.ManagerName, objects []string) error {
	pm := m.pmPnt(pmName)
	switch oType {
	case TypePackage:
		pm.Global.Packages = lo.Filter(pm.Global.Packages, func(item string, index int) bool {
			return !lo.Contains(objects, item)
		})
	case TypeDependency:
		pm.Global.Dependencies = lo.Filter(pm.Global.Dependencies, func(item string, index int) bool {
			return !lo.Contains(objects, item)
		})
	}
	return nil
}

// FIXME: Somhow os.Hostname needs to be moved to an interface
func (m *Manifest) AddToHost(toAdd []string, pmName shared.ManagerName, oType ManifestObjectType) error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	return m.AddConditional(oType, pmName, MConditionHost, hostname, toAdd)
}

func (m *Manifest) AddToGroup(toAdd []string, group string, pmName shared.ManagerName, oType ManifestObjectType) error {
	return m.AddConditional(oType, pmName, MConditionGroup, group, toAdd)
}

func (m *Manifest) getOrAddConditional(pmName shared.ManagerName, cType ManifestConditionalType, cValue string) (*Conditional, error) {
	pm := m.pmPnt(pmName)
	for idx := range pm.Conditional {
		if pm.Conditional[idx].Type == cType && pm.Conditional[idx].Value == cValue {
			return &pm.Conditional[idx], nil
		}
	}
	pm.Conditional = append(pm.Conditional, Conditional{
		Type:         cType,
		Value:        cValue,
		Dependencies: []string{},
		Packages:     []string{},
	})
	return &pm.Conditional[len(pm.Conditional)-1], nil
}

func (m *Manifest) pmPnt(name shared.ManagerName) *PmManifest {
	switch name {
	case "dnf":
		return &m.Dnf
	case "git":
		return &m.Git
	case "go":
		return &m.Go
	default:
		panic(fmt.Sprintf("%s is not a registered package manager", name))
	}
}

func (m *Manifest) Save(filename string) error {
	m.Version = config.Version
	var b bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&b)
	yamlEncoder.SetIndent(2)
	err := yamlEncoder.Encode(&m)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, b.Bytes(), 0755)
}

type PmManifest struct {
	Global      Global        `yaml:"global"`
	Conditional []Conditional `yaml:"conditional"`
}

type Global struct {
	Dependencies []string `yaml:"dependencies"`
	Packages     []string `yaml:"packages"`
}

type Conditional struct {
	Type         ManifestConditionalType `yaml:"type"`
	Value        string                  `yaml:"value"`
	Dependencies []string                `yaml:"dependencies"`
	Packages     []string                `yaml:"packages"`
}

func InitManifest() (Manifest, error) {
	return readManifest(config.ManifestFile)
}

func readManifest(filename string) (manifest Manifest, err error) {
	err = createOrMigrateManifestFile(filename)
	if err != nil {
		return
	}

	yamlRaw, err := os.ReadFile(filename)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(yamlRaw, &manifest)
	if err != nil {
		return
	}
	return
}

func createOrMigrateManifestFile(filename string) error {
	// err := os.MkdirAll(ConfigDir, os.ModePerm)
	// if err != nil {
	// 	return err
	// }

	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return createManifestFile(filename)
	}

	if info.IsDir() {
		return fmt.Errorf("%s is a directory", filename)
	}

	return nil
}

func createManifestFile(filename string) error {
	bytes, err := yaml.Marshal(Manifest{})
	if err != nil {
		return err
	}
	return os.WriteFile(filename, bytes, 0755)
}
