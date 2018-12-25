#!/usr/bin/env sh

#if [ -z  "$1" ]; then
    export PREFIX=/usr
#else
#    export PREFIX=$1
#fi

if [ "$PREFIX" = "/usr" ] && [ "$(id -u)" != "0" ]; then
    # Make sure only root can run our script
    echo "This script must be run as root" 1>&2
    exit 1
fi

# Check availability of required commands
# COMMANDS="install glib-compile-schemas glib-compile-resources msgfmt desktop-file-validate gtk-update-icon-cache"
COMMANDS="install glib-compile-schemas glib-compile-resources msgfmt desktop-file-validate gtk-update-icon-cache"
# if [ "$PREFIX" = '/usr' ] || [ "$PREFIX" = "/usr/local" ]; then
#     COMMANDS="$COMMANDS xdg-desktop-menu"
# fi
# PACKAGES="coreutils glib2 glib2 gettext desktop-file-utils gtk-update-icon-cache xdg-utils"
PACKAGES="coreutils glib2 glib2 gettext desktop-file-utils gtk-update-icon-cache xdg-utils"
i=0
for COMMAND in $COMMANDS; do
    type $COMMAND >/dev/null 2>&1 || {
        j=0
        for PACKAGE in $PACKAGES; do
            if [ $i = $j ]; then
                break
            fi
            j=$(( $j + 1 ))
        done
        echo "Your system is missing command $COMMAND, please install $PACKAGE"
        exit 1
    }
    i=$(( $i + 1 ))
done

SCRIPT_DIR=$(dirname "$0")

echo "Installing gsettings schema to prefix ${PREFIX}"

# Copy and compile schema
echo "Copying and compiling schema..."
install -d ${PREFIX}/share/glib-2.0/schemas
install -m 644 ${SCRIPT_DIR}/gsettings/org.d2r2.gorsync.gschema.xml ${PREFIX}/share/glib-2.0/schemas/
glib-compile-schemas ${PREFIX}/share/glib-2.0/schemas/

