#!/usr/bin/env bash
#
# Example showing use of getopt detection and use of GNU enhanced getopt
# to handle arguments containing whitespace.
#
# Written in 2004 by Hoylen Sue <hoylen@hoylen.com>
# Modified in 2018 by Denis Dyakov <denis.dyakov@gmail.com>
#
# To the extent possible under law, the author(s) have dedicated all copyright and
# related and neighboring rights to this software to the public domain worldwide.
# This software is distributed without any warranty.
#
# You should have received a copy of the CC0 Public Domain Dedication along with this software.
# If not, see <http://creativecommons.org/publicdomain/zero/1.0/>.

# Install in advance 'fmp' application according to: https://fpm.readthedocs.io/en/latest/installing.html

# !!! This script is a part of distribution packaging system !!!
# !!! Script work together with  gs_schema_install.sh/..._uninstall.sh to [de]install app GLIB schema file !!!
# !!! Change with great care, do not break it !!!

# Get desktop file content to register application in
# Linux GUI environment.
get_desktop_entry_file()
{
    local EMBEDDED
    EMBEDDED=$(cat << EndOfMsg
[Desktop Entry]
Name=Gorsync Backup
Comment=Easy-to-use backup app based on Rsync console utility
Exec=gorsync
Icon=media-tape-symbolic
Type=Application
Encoding=UTF-8
Terminal=false
Categories=GNOME;GTK;Utility;System;
Keywords=backup;rsync;
EndOfMsg
)
    echo "${EMBEDDED}"
}

PROG=$(basename $0)
VERSION=v0.3

# Remove this trap if you are doing your own error detection or don't care about errors
trap "echo $PROG: error encountered: aborted; exit 3" ERR

#----------------------------------------------------------------
# Process command line arguments

## Define options: trailing colon means has an argument (customize this: 1 of 3)

SHORT_OPTS=h
LONG_OPTS=version,skip-app-build,help,version

SHORT_HELP="Usage: ${PROG} [options] arguments
Options:
  -h                        Show this help message."

LONG_HELP="Usage: ${PROG} [options] arguments
Options:
  --help                         Show this help message.
  --skip-app-build               Skip compiling of application itself.
  --version                      Show version information."

# Detect if GNU Enhanced getopt is available

HAS_GNU_ENHANCED_GETOPT=
if getopt -T >/dev/null; then :
else
  if [ $? -eq 4 ]; then
    HAS_GNU_ENHANCED_GETOPT=yes
  fi
fi

# Run getopt (runs getopt first in `if` so `trap ERR` does not interfere)

if [ -n "$HAS_GNU_ENHANCED_GETOPT" ]; then
  # Use GNU enhanced getopt
  if ! getopt --name "$PROG" --long $LONG_OPTS --options $SHORT_OPTS -- "$@" >/dev/null; then
    echo "$PROG: usage error (use -h or --help for help)" >&2
    exit 2
  fi
  ARGS=$(getopt --name "$PROG" --long $LONG_OPTS --options $SHORT_OPTS -- "$@")
else
  # Use original getopt (no long option names, no whitespace, no sorting)
  if ! getopt $SHORT_OPTS "$@" >/dev/null; then
    echo "$PROG: usage error (use -h for help)" >&2
    exit 2
  fi    
  ARGS=$(getopt $SHORT_OPTS "$@")
fi
eval set -- $ARGS

## Process parsed options (customize this: 2 of 3)
 
while [ $# -gt 0 ]; do
    case "$1" in
        --skip-app-build)   SKIP_APP_BUILD=true;;
        -h | --help)     if [ -n "$HAS_GNU_ENHANCED_GETOPT" ]
                         then echo "$LONG_HELP";
                         else echo "$SHORT_HELP";
                         fi;  exit 1;;
        --version)       echo "$PROG $VERSION"; exit 1;;
        --)              shift; break;; # end of options
    esac
    shift
done

echo "************************************************************************"
echo "This script will generate linux distribution packages using fpm utility."
echo "You must install fpm in advance to use this script:"
echo "https://fpm.readthedocs.io/en/latest/installing.html"
echo "************************************************************************"


TEMPDIR=/tmp/gorsync_build_app
DISTRIB=distrib
SCRIPTS=scripts
APP_NAME='gorsync'
APP_URL="https://gorsync.github.io"
AUTHOR="Denis Dyakov <denis.dyakov@gmail.com>"
LICENSE="GPL3"

echo -e "Use path to keep data for package generation: ${TEMPDIR}\n"

systems=( \
    # for Archlinux
    "ARCHLINUX" \
    # for Debian, Ubuntu
    "DEBIAN" \
    # for Redhat, Centos
    "REDHAT" \
    # for FreeBSD
    "FREEBSD")
prefixes=( \
    # for Archlinux
    "usr" \
    # for Debian, Ubuntu
    "usr" \
    # for Redhat, Centos
    "usr" \
    # for FreeBSD
    "usr/local")
fpm_packages=( \
    # for Archlinux
    "pacman" \
    # for Debian, Ubuntu
    "deb" \
    # for Redhat, Centos
    "rpm" \
    # for FreeBSD
    "freebsd")
warnings=( \
    # for Archlinux
    "" \
    # for Debian, Ubuntu
    "" \
    # for Redhat, Centos
    "" \
    # for FreeBSD
    "Application binary '${APP_NAME}' generated on Linux is not valid for FreeBSD.\n\tBuild it natively on FreeBSD, substitute and re-run with --skip-app-build option.")
fpm_dependencies=( \
    # for Archlinux
    "--depends rsync --depends glib2 --depends gtk3 --depends libnotify" \
    # for Debian, Ubuntu
    "--depends rsync --depends libglib2.0-dev --depends libgtk-3-dev --depends libnotify-dev" \
    # for Redhat, Centos
    "--depends rsync --depends glib2-devel --depends gtk3 --depends libnotify-devel" \
    # for FreeBSD
    "--depends rsync --depends glib --depends gtk3 --depends libnotify")

# rm -R $TEMPDIR >/dev/null 2>&1

for ((i=0; i<${#systems[@]};++i))
do
    echo "Start packaging ${systems[i]}..."

    mkdir -p $TEMPDIR/${systems[i]}/$DISTRIB/${prefixes[i]}/bin
    mkdir -p $TEMPDIR/${systems[i]}/$DISTRIB/${prefixes[i]}/share/applications
    mkdir -p $TEMPDIR/${systems[i]}/$SCRIPTS

    SAVE_DIR="${PWD}"
    cd ../..
    PARENT_DIR="${PWD}"

    cp "$PARENT_DIR/ui/gtkui/gs_schema_install.sh" "$TEMPDIR/${systems[i]}/$SCRIPTS"
    # Prepare and embed xml file as HEREDOC into the gs_schema_install.sh
    XML_SCHEMA=$(cat $PARENT_DIR/ui/gtkui/gsettings/org.d2r2.gorsync.gschema.xml)
    XML_SCHEMA="${XML_SCHEMA//\\/\\\\}"
    XML_SCHEMA="${XML_SCHEMA//\//\\/}"
    XML_SCHEMA="${XML_SCHEMA//&/\\&}"
    XML_SCHEMA="${XML_SCHEMA//$'\n'/\\n}"
    sed -i "s/# AUTOMATICALLY_REPLACED_WITH_EMBEDDED_XML_FILE_DECLARATION/EMBEDDED=$\(cat << EndOfMsg\n${XML_SCHEMA}\nEndOfMsg\n)/" \
        "$TEMPDIR/${systems[i]}/$SCRIPTS/gs_schema_install.sh"
    cp "$PARENT_DIR/ui/gtkui/gs_schema_uninstall.sh" "$TEMPDIR/${systems[i]}/$SCRIPTS"

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
    ITERATION=$(($COMMITS_AFTER+1))

    APP_BUILD_SUCCESSFULL=true
    if [ -z $SKIP_APP_BUILD ]; then
        ./gorsync_build.sh --buildtype Release
        if [ $? -eq 0 ]; then
            echo "App successfully compiled."
            cd "$SAVE_DIR"
            cp "$PARENT_DIR/$APP_NAME" "$TEMPDIR/${systems[i]}/$DISTRIB/${prefixes[i]}/bin"
        else
            APP_BUILD_SUCCESSFULL=false
        fi
    fi
    cd "$SAVE_DIR"
    echo "$(get_desktop_entry_file)" > "$TEMPDIR/${systems[i]}/$DISTRIB/${prefixes[i]}/share/applications/gorsync.desktop"

    if [ $APP_BUILD_SUCCESSFULL = true ]; then

        mkdir -p ./packages && cd ./packages

        fpm -s dir -f \
            -t ${fpm_packages[i]} \
            -C "$TEMPDIR/${systems[i]}/$DISTRIB" \
            --name $APP_NAME \
            --version $APP_VERSION \
            --iteration $ITERATION  \
            --description "GTK+ frontend (backup application) for RSYNC utility" \
            ${fpm_dependencies[i]} \
            --after-install "$TEMPDIR/${systems[i]}/$SCRIPTS/gs_schema_install.sh" \
            --before-remove "$TEMPDIR/${systems[i]}/$SCRIPTS/gs_schema_uninstall.sh" \
            --conflicts "gorsync-git" \
            --maintainer "$AUTHOR" \
            --url "$APP_URL" \
            --license "$LICENSE"
        #    --config-files /etc

        # Replacing single-spaces with empty via contruction ${MYSTRING//in/by} (or) ${MYSTRING//in}
        if [ -z ${warnings[i]// } ]; then
            :
        else
            echo -e "WARNING: ${warnings[i]}"
        fi

        echo -e "...${systems[i]} done.\n"
    else
        echo -e "...${systems[i]} FAIL.\n"
    fi
    cd "$SAVE_DIR"

done

