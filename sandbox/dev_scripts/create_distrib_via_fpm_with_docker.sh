#!/usr/bin/env bash
# This script is experimental and might not work correctly!
# Was created as an option to build 32bit packages on 64bit OS.
# Not sure about results.

DKR="sudo docker"
LINUX_64BIT=debian
LINUX_32BIT=i386/debian
CNR_64BIT=gorsync_build_64bit
CNR_32BIT=gorsync_build_32bit
GOARCH_64BIT=amd64
GOARCH_32BIT=386


deploy_docker_container() {
    local CNR_NAME=$1
    local DISTR_NAME=$2

    $DKR ps | grep $CNR_NAME && $DKR stop $CNR_NAME
    $DKR ps -a | grep $CNR_NAME && $DKR rm $CNR_NAME
    $DKR create --name=$CNR_NAME -it $DISTR_NAME && $DKR start $CNR_NAME
    $DKR exec $CNR_NAME apt update && $DKR exec $CNR_NAME apt upgrade -y
    $DKR exec $CNR_NAME apt install -y curl wget
}

install_dependencies() {
    local CNR_NAME=$1

    $DKR exec $CNR_NAME apt install -y rsync libglib2.0-dev libgtk-3-dev libnotify-dev git ruby ruby-dev rubygems build-essential bsdtar rpm
    $DKR exec $CNR_NAME gem install --no-ri --no-rdoc fpm
}

install_golang() {
    local CNR_NAME=$1
    local GOARCH=$2

    # $DKR exec $CNR_NAME curl -sL -o /usr/local/sbin/gimme https://raw.githubusercontent.com/travis-ci/gimme/master/gimme
    # $DKR exec $CNR_NAME chmod +x /usr/local/sbin/gimme
    # $DKR exec $CNR_NAME sh -c 'GIMME_ARCH=$GOARCH gimme 1.13 | grep "GOROOT"'
    $DKR exec $CNR_NAME apt install -y golang
}

build_gorsync_packages() {
    local CNR_NAME=$1

    GOCODEPATH=/root/Downloads/gocode
    GORSYNCSUBPATH=github.com/d2r2/go-rsync
    PACKAGINGSUBPATH=packaging/fpm_packages
    # GOROOTPATH=$($DKR exec $CNR_NAME gimme 1.13 | grep "GOROOT" | sed "s/^.*'\(.*\)'.*$/\1/")
    echo $GOROOTPATH
    $DKR exec $CNR_NAME sh -c "GOROOT=$GOROOTPATH PATH=\$GOROOT/bin:\$PATH GOPATH=$GOCODEPATH go version"
    $DKR exec $CNR_NAME mkdir -p $GOCODEPATH
    $DKR exec $CNR_NAME sh -c "GOROOT=$GOROOTPATH PATH=\$GOROOT/bin:\$PATH GOPATH=$GOCODEPATH go get -u -v github.com/d2r2/go-rsync"
    $DKR exec $CNR_NAME sh -c "GOROOT=$GOROOTPATH PATH=\$GOROOT/bin:\$PATH GOPATH=$GOCODEPATH go get -u -v all"
    $DKR exec $CNR_NAME sh -c "cd $GOCODEPATH/src/$GORSYNCSUBPATH/$PACKAGINGSUBPATH && GOROOT=$GOROOTPATH PATH=\$GOROOT/bin:\$PATH GOPATH=$GOCODEPATH ./create_distrib_packages_with_fpm.sh"
}

# deploy and build 64 bit Gorsync packages
# deploy_docker_container $CNR_64BIT $LINUX_64BIT
# install_dependencies $CNR_64BIT
# build_gorsync_packages $CNR_64BIT

# deploy and build 32 bit Gorsync packages
deploy_docker_container $CNR_32BIT $LINUX_32BIT
install_dependencies $CNR_32BIT
install_golang $CNR_32BIT $GOARCH_32BIT
build_gorsync_packages $CNR_32BIT


