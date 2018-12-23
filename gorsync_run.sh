#!/usr/bin/env bash
./gorsync_build.sh $@
[ $? -eq 0 ] && ./gorsync || exit $?
