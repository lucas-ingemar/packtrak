<p align="center">
    <picture>
      <img alt="packtrak" title="packtrak" src="docs/assets/img/logowtext.png">
    </picture>
</p>

![GitHub stars](https://img.shields.io/github/stars/lucas-ingemar/packtrak.svg?label=github%20stars)
![Latest GitHub release](https://img.shields.io/github/release/lucas-ingemar/packtrak.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/lucas-ingemar/packtrak)](https://goreportcard.com/report/github.com/lucas-ingemar/packtrak)

## Track Your Package Managers
You have all your dotfiles in sync between your systems and you know you can set up a new system instantly. You get a new computer, download your dotfiles and think you are ready to go. But now you remember - I need to install a lot of system packages to get my setup working...
Now there is a solution to this problem! Packtrak tracks your package managers and keeps a manifest containing all the packages you want your system to have. 
The manifest is easily revisioned and can just like your dotfiles be shared across systems.

## Install

### Go 
Use go install to compile locally. 

``` bash
go install github.com/lucas-ingemar/packtrak/cmd/packtrak@latest
```

## Quick Usage Guide

Install a package:
``` bash
packtrak [manager] install [package]
```

Remove a package:
``` bash
packtrak [manager] remove [package]
```

Install a depencdency:
``` bash
packtrak [manager] install --dependency [dependency]
```

Remove a dependency:
``` bash
packtrak [manager] remove --dependency [dependency]
```

Synchronize manifest and system: 
``` bash
packtrak sync 
```

See the [documentation](docs/cmd/packtrak.md) for more information.


## Autocompletion
Packtrak generates its own autocompletion for the commands. Simply put the following command in your `.bashrc`, `.zshrc` or the corresponding file for your setup:

``` bash
source <(./build/packtrak completion [shell])
```

## Manifest File
The wanted package state for the managers is listed in `~/.config/packtrak/manifest.yaml`. This file can be shared between computers to sync the state between them. When installing and removing packages this file will be updated.

An example of the file layout:

``` yaml
dnf:
  global:
    dependencies:
      - copr:copr.fedorainfracloud.org/phracek/PyCharm
    packages:
      - sway
      - kanshi
      - btop
  conditional:
    - type: group
      value: work
      dependencies: 
        - cm:https://rpm.releases.hashicorp.com/fedora/hashicorp.repo
      packages: 
        - terraform
    - type: group
      value: home
      dependencies: []
      packages:
        - htop
    - type: host
      value: spock
      dependencies: []
      packages: 
        - krita
git:
  global:
    dependencies: []
    packages:
      - https://github.com/zsh-users/zsh-autosuggestions.git
      - https://github.com/binpash/try.git
      - https://github.com/ahmetb/kubectx
  conditional: []
go:
  global:
    dependencies: []
    packages:
      - github.com/lucas-ingemar/packtrak/cmd/packtrak
      - github.com/mikefarah/yq/v4
      - github.com/lucas-ingemar/clergo/cmd/clergo
      - github.com/golangci/golangci-lint/cmd/golangci-lint
      - golang.org/x/tools/gopls
      - sigs.k8s.io/kind
      - github.com/rogpeppe/godef
  conditional: []
_version: v0.9.0

```
### Global
Everything under `global` will be installed on all machines with this file. This should contain the all the packages that you want to share between all your systems. 
If a package or dependency is installed without any arguments they will be added under the global category.

### Conditional
Under `conditional` different rules can be applied. These rules are used to only install packages or dependencies on systems that match the rules.

#### Group
The group is matched against what is defined in the config file under `groups`. If the group in the manifest matches one of the groups in the config file the rule is applied.

#### Host
The host matches the hostname of the system. If it is a match the rule is applied.

