#!/usr/bin/env sh

#### Create folders
mkdir -p /testing_dir/git
mkdir -p /testing_dir/github/library
mkdir -p /testing_dir/github/bin
mkdir -p /testing_dir/go/bin

#### DNF #########
CMD_FOUND=$(command -v bat)
if [ "$CMD_FOUND" != "" ]; then
    echo "ERROR: dnf command already exists"
    exit 1
fi

go run cmd/packtrak/main.go -y dnf install bat

CMD_FOUND=$(command -v bat)
if [ "$CMD_FOUND" == "" ]; then
    echo "ERROR: dnf command did not install"
    exit 1
fi

go run cmd/packtrak/main.go -y dnf remove bat

CMD_FOUND=$(command -v bat)
if [ "$CMD_FOUND" != "" ]; then
    echo "ERROR: dnf command did not uninstall"
    exit 1
fi

##### GO ########
export GOPATH=/testing_dir/go

ls /testing_dir/go/bin/packtrak &> /dev/null
if [[ $? == 0 ]]; then
    echo "ERROR: go command already exists"
    exit 1
fi

go run cmd/packtrak/main.go -y go install github.com/lucas-ingemar/packtrak/cmd/packtrak

ls /testing_dir/go/bin/packtrak &> /dev/null
if [[ $? != 0 ]]; then
    echo "ERROR: go command did not install"
    exit 1
fi

go run cmd/packtrak/main.go -y go remove packtrak

ls /testing_dir/go/bin/packtrak &> /dev/null
if [[ $? == 0 ]]; then
    echo "ERROR: go command did not uninstall"
    exit 1
fi

#### GIT #########
ls /testing_dir/git/lucas-ingemar.packtrak &> /dev/null
if [[ $? == 0 ]]; then
    echo "ERROR: git command already exists"
    exit 1
fi

go run cmd/packtrak/main.go -y git install https://github.com/lucas-ingemar/packtrak.git

ls /testing_dir/git/lucas-ingemar.packtrak &> /dev/null
if [[ $? != 0 ]]; then
    echo "ERROR: git command did not install"
    exit 1
fi

go run cmd/packtrak/main.go -y git remove lucas-ingemar/packtrak

ls /testing_dir/git/lucas-ingemar.packtrak &> /dev/null
if [[ $? == 0 ]]; then
    echo "ERROR: git command did not uninstall"
    exit 1
fi

#### GITHUB #########
####
ls /testing_dir/github/library/ahmetb.kubectx* &> /dev/null
if [[ $? == 0 ]]; then
    echo "ERROR: github command already exists"
    exit 1
fi

ls /testing_dir/github/bin/kubectx &> /dev/null
if [[ $? == 0 ]]; then
    echo "ERROR: github bin symlink already exists"
    exit 1
fi

go run cmd/packtrak/main.go -y github install github.com/ahmetb/kubectx:kubectx_#version#_linux_x86_64.tar.gz

ls /testing_dir/github/library/ahmetb.kubectx* &> /dev/null
if [[ $? != 0 ]]; then
    echo "ERROR: github command did not install"
    exit 1
fi

ls /testing_dir/github/bin/kubectx &> /dev/null
if [[ $? != 0 ]]; then
    echo "ERROR: github bin symlink did not install"
    exit 1
fi

go run cmd/packtrak/main.go -y github remove ahmetb/kubectx

ls /testing_dir/github/library/ahmetb.kubectx* &> /dev/null
if [[ $? == 0 ]]; then
    echo "ERROR: github command did not uninstall"
    exit 1
fi

ls /testing_dir/github/bin/kubectx &> /dev/null
if [[ $? == 0 ]]; then
    echo "ERROR: github bin symlink did not uninstall"
    exit 1
fi

#### FLATPAK - Not testet at the moment. Too massive deps are downloaded
# flatpak list --system | grep peg-e &> /dev/null
# if [[ $? == 0 ]]; then
#     echo "ERROR: flatpak command already exists"
#     exit 1
# fi

# go run cmd/packtrak/main.go -y flatpak install flathub:org.gottcode.Peg-E

# flatpak list --system | grep peg-e &> /dev/null
# if [[ $? != 0 ]]; then
#     echo "ERROR: flatpak command did not install"
#     exit 1
# fi
