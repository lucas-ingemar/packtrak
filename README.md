![GitHub stars](https://img.shields.io/github/stars/lucas-ingemar/packtrak.svg?label=github%20stars)
![Latest GitHub release](https://img.shields.io/github/release/lucas-ingemar/packtrak.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/lucas-ingemar/packtrak)](https://goreportcard.com/report/github.com/lucas-ingemar/packtrak)

# packtrak 
Track your packages

## State - PLAN

Use sqlite. Only save whats installed: Which package manager + the "real" app-name. For dnf it would be `btop`, but for go it would be: `golang.org/x/tools/gopls`. 
Don't care about version tags, they doesn't matter for the state.

So, the packages in the state is one table that looks something like this:

```
id        manager        package
-----------------------------------------
0         dnf            btop
1         go             golang.org/x/tools/gopls
2         git            git@github.com:lucas-ingemar/packtrak.git
```

## Package file - PLAN
Should only have 1 field: `package`. This field will have the same convention as `package` in the state.
All the logic should be happening in the different managers inside the app. E.g: go package renaming from `golang.org/x/tools/gopls` to `gopls` in the UI.

``` yaml
dnf:
  global:
    dependencies:
      - copr:copr.fedorainfracloud.org/phracek/PyCharm
      - cm:https://rpm.releases.hashicorp.com/fedora/hashicorp.repo
    packages:
      - sway
      - kanshi
      - btop
      - terraform
  conditional:
    - type: host
      value: spock
      dependencies:
        - hejhej
      packages:
        - bla
    - type: group
      value: work
      dependencies:
        - dada
      packages:
        - apa
go:
  global:
    packages:
      - github.com/mikefarah/yq/v4
      - github.com/lucas-ingemar/clergo/cmd/clergo
      - github.com/rogpeppe/godef
      - github.com/golangci/golangci-lint/cmd/golangci-lint
      - golang.org/x/tools/gopls
      - sigs.k8s.io/kind
```

### Host/environment specific
~~Should add rules next to the global packages. For instance for host-specific packages, or `group(?): work` packages~~

## Package versions
To start with only `latest` will be installed

## Manager specific setttings.
Take dnf for example. Should be able to list coprs that should be installed. That needs to be solved
Maybe there could be something like dependencies for each manager as well
That have its own table in the state.

# TODO:
* ~~Should probably commit db for each packagemanager in sync as well~~
* ~~Rotate state file. Config option for how many rotations~~
* ~~Make sure user is not sudo when cmd runs~~
* ~~Add dependencies. Example: COPR/external sources for dnf~~
* Add update function ??
* ~~Handle add/remove commands with conditional~~
* Custom config variables
* Add initCheck to make sure special config is correct
* ~~Add initCheckCommand to make sure the correct command exists on the host~~
* Fix all cmd descriptions and such

# KNOWN ISSUES:
* ~~Must trigger sudo early on, so that dnf is running under sudo without password~~
