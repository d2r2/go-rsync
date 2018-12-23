#!/usr/bin/env sh

#if [ -z  "$1" ]; then
    export PREFIX=/usr
    # Make sure only root can run our script
    if [ "$(id -u)" != "0" ]; then
        echo "This script must be run as root" 1>&2
        exit 1
    fi
#else
#    export PREFIX=$1
#fi

echo "Uninstalling gsettings schema from prefix ${PREFIX}"

rm ${PREFIX}/share/glib-2.0/schemas/org.d2r2.gorsync.gschema.xml
glib-compile-schemas ${PREFIX}/share/glib-2.0/schemas/
