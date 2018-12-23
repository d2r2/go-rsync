#!/usr/bin/env bash

# Read manuals to understand how to build aur package in Arch Linux:
#	https://wiki.archlinux.org/index.php/Arch_User_Repository
#	https://wiki.archlinux.org/index.php/Arch_package_guidelines
#	https://wiki.archlinux.org/index.php/Creating_packages
#	https://wiki.archlinux.org/index.php/PKGBUILD
#	https://wiki.archlinux.org/index.php/Makepkg
# Examples aur packages with go sources:
#	https://aur.archlinux.org/cgit/aur.git/tree/PKGBUILD?h=gometalinter-git
#	https://aur.archlinux.org/cgit/aur.git/tree/PKGBUILD?h=vim-go
#	https://aur.archlinux.org/cgit/aur.git/tree/PKGBUILD?h=gotags-git

TEMP_DIR=$(mktemp -d)
SAVE_DIR="${PWD}"

git clone ssh://aur@aur.archlinux.org/gorsync-git.git "${TEMP_DIR}"
#cp -R ./gorsync-git $TEMP_DIR
cp ./gorsync-git/PKGBUILD "${TEMP_DIR}/"
cp ./gorsync-git/gorsync-git.install "${TEMP_DIR}/"
cd "${TEMP_DIR}"
makepkg --printsrcinfo > .SRCINFO
echo "go to ${TEMP_DIR} and run makepkg..."
