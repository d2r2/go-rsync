#!/usr/bin/env bash

# !!! This script is a part of distribution packaging system !!!
# !!! Change with great care, do not break it !!!


get_make_shell_script()
{
    local EMBEDDED
    EMBEDDED=$(cat << EndOfMsg
#!/usr/bin/env sh
make -C ./src/github.com/d2r2/go-rsync
EndOfMsg
)
    echo "${EMBEDDED}"
}


echo "****************************************************************************************************"
echo "This script will convert project sources to source archive file, embedding all external dependencies"
echo "into the single file. This will preserve application sources release consistency and allow to compile"
echo "application binary at any time later, without necessity to download dependencies."
echo "****************************************************************************************************"

APP_NAME='gorsync'
# Form application version from latest GIT tag/release.
# Extract latest GIT tag.
GIT_TAG=$(git describe --tags --abbrev=0)
# Extract number of commits passed from last GIT release.
COMMITS_AFTER=$(git rev-list ${GIT_TAG}..HEAD --count)
# Remove 'v' char from tag, if present
[[ ${GIT_TAG:0:1} == "v" ]] && GIT_TAG=${GIT_TAG:1}
# Combine last GIT tag and number of commits since, if applicable, to build application version.
APP_VERSION=$GIT_TAG
# Add extra 1 to increment build number (to start index from 1).
[[ "$COMMITS_AFTER" != "0" ]] && APP_VERSION="$GIT_TAG-$(($COMMITS_AFTER+1))"


TEMP_DIR=$(mktemp -d)

CURDIR=$PWD
# echo ${CURDIR}
PROJECT_PATH="${CURDIR%/*/*}"
# echo $PROJECT_PATH
GOCODE_PATH="${PROJECT_PATH%/*/*/*/*}"

PROJECT_SUBPATH="${PROJECT_PATH##$GOCODE_PATH}"
# echo $PROJECT_SUBPATH
GOCODE_SRC_SUBPATH="${PROJECT_SUBPATH%/*}"
# echo $GOCODE_SRC_SUBPATH

echo "Copying project sources to: ${TEMP_DIR}${GOCODE_SRC_SUBPATH}..."
mkdir -p $TEMP_DIR$GOCODE_SRC_SUBPATH
rsync  -avrq --exclude packaging/build_packages/packages $PROJECT_PATH $TEMP_DIR$GOCODE_SRC_SUBPATH
# cp ./Makefile $TEMP_DIR$PROJECT_SUBPATH
SCRIPT_NAME=make_app_from_archive_source.sh
echo "$(get_make_shell_script)" > ${TEMP_DIR}/${SCRIPT_NAME}
chmod +x ${TEMP_DIR}/${SCRIPT_NAME}
rm -R $TEMP_DIR$PROJECT_SUBPATH/packaging/build_packages/packages 2>/dev/null
rm $TEMP_DIR$PROJECT_SUBPATH/gorsync 2>/dev/null
# echo $TEMP_DIR$GOCODE_SRC_SUBPATH

GOPATH=$TEMP_DIR
GOBIN=$TEMP_DIR/bin
echo "Installing govendor tool..."
GOVENDORURL=github.com/kardianos/govendor
go get -u $GOVENDORURL
go install $GOVENDORURL
# echo $TEMP_DIR$PROJECT_SUBPATH
cd $TEMP_DIR$PROJECT_SUBPATH
$GOBIN/govendor init
# $GOBIN/govendor list # uncomment for debugging purpose
echo "Converting project missing or external packages to embedded vendor packages, using govendor tool..."
$GOBIN/govendor fetch -v +outside
cd $CURDIR
echo "Removing govendor tool..."
go clean -i $GOVENDORURL
rm -R --force ${TEMP_DIR}/src/$(dirname $GOVENDORURL)

for dir in $TEMP_DIR/*/ $TEMP_DIR/.[^.]*/ ; do
    # echo "$dir"
    last_dir=$(basename $dir)
    if [ "$last_dir" != "src" ]; then
        rm -R --force $dir
    fi
done


ARCHIVE_PATH="./packages"
ARCHIVE_SOURCE=${APP_NAME}_${APP_VERSION}.tar.gz
echo "Building archive ${ARCHIVE_PATH}/${ARCHIVE_SOURCE}..."
mkdir -p ./packages
rm ./packages/${ARCHIVE_SOURCE} 2>/dev/null
tar cfz ./packages/${ARCHIVE_SOURCE} -C ${TEMP_DIR} ./

echo "...done."
