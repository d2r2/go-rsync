#!/usr/bin/env bash
#
# Example showing use of getopt detection and use of GNU enhanced getopt
# to handle arguments containing whitespace.
#
# Written in 2004 by Hoylen Sue <hoylen@hoylen.com>
# Modified in 2018-2020 by Denis Dyakov <denis.dyakov@gmail.com>
#
# To the extent possible under law, the author(s) have dedicated all copyright and
# related and neighboring rights to this software to the public domain worldwide.
# This software is distributed without any warranty.
#
# You should have received a copy of the CC0 Public Domain Dedication along with this software.
# If not, see <http://creativecommons.org/publicdomain/zero/1.0/>.

PROG=$(basename $0)
VERSION=v0.4

# Define default values, if parameters not specified
RELEASE_TYPE="Release"
DEV_TYPE="Development"

# Remove this trap if you are doing your own error detection or don't care about errors
trap "echo $PROG: error encountered: aborted; exit 3" ERR

#----------------------------------------------------------------
# Process command line arguments

## Define options: trailing colon means has an argument (customize this: 1 of 3)

SHORT_OPTS=b:t:o:h
LONG_OPTS=buildtype:,tags:,output:,version,race,help

SHORT_HELP="Usage: ${PROG} [options] arguments
Options:
  -b <build type>           Build type. Release type = ${RELEASE_TYPE}.
  -t <golang tags>          Build tags.
  -o <output>               Specify output path to write app binary.
  -h                        Show this help message."

LONG_HELP="Usage: ${PROG} [options] arguments
Options:
  -b | --buildtype <build type>       Build type. Release type = ${RELEASE_TYPE}.
  -t | --tags <golang tags>           Build tags.
  -o | --output <output>              Specify output path to write app binary.
  -h | --help                         Show this help message.
  -r | --race                         Investigate application race conditions.
  --version                           Show version information."

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
        -b | --buildtype)   BUILD_TYPE="$2"; shift;;
        -t | --tags)        BUILD_TAGS="$2"; shift;;
        -o | --output)      OUTPUT="-o $2"; shift;;
        -v | --verbose)     VERBOSE=true;;
        -r | --race)        RACE="-race";;
        -h | --help)     if [ -n "$HAS_GNU_ENHANCED_GETOPT" ]
                         then echo "$LONG_HELP";
                         else echo "$SHORT_HELP";
                         fi;  exit 1;;
        --version)       echo "$PROG $VERSION"; exit 1;;
        --)              shift; break;; # end of options
    esac
    shift
done

# Form application version from latest GIT tag/release.
# Extract latest GIT tag.
GIT_TAG=$(git describe --tags --abbrev=0)
# Extract number of commits passed from last GIT release.
COMMITS_AFTER=$(git rev-list ${GIT_TAG}..HEAD --count)
COMMIT_ID=$(git rev-parse --short=7 HEAD)
# Remove 'v' char from tag, if present
[[ ${GIT_TAG:0:1} == "v" ]] && GIT_TAG=${GIT_TAG:1}
# Combine last GIT tag and number of commits since, if applicable, to build application version.
APP_VERSION=$GIT_TAG
[[ "$COMMITS_AFTER" != "0" ]] && APP_VERSION="$GIT_TAG+$(($COMMITS_AFTER))~g$COMMIT_ID"

shopt -s nocasematch
if [[ "$BUILD_TYPE" == "$RELEASE_TYPE" ]]; then
  echo "RELEASE type build in progress..."
  go run data/generate/generate.go && mv ./assets_vfsdata.go ./data
  # Add extra options here (-s -w), to decrease release binary size, read here https://golang.org/cmd/link/
  go build -v $RACE -ldflags="-X main.version=$APP_VERSION -X main.buildnum=$(date -u +%Y%m%d%H%M%S) -s -w" -tags "gorsync_rel $BUILD_TAGS" $OUTPUT gorsync.go
else
  [[ -z "$BUILD_TYPE" ]] || [[ "$BUILD_TYPE" == "$DEV_TYPE" ]] || echo "WARNING: unknown build type provided: $BUILD_TYPE"
  echo "DEVELOPMENT type build in progress..."
  go build -v $RACE -ldflags="-X main.version=$APP_VERSION -X main.buildnum=$(date -u +%Y%m%d%H%M%S)" -tags "$BUILD_TAGS" $OUTPUT gorsync.go
fi
shopt -u nocasematch

