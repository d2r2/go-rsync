#!/usr/bin/env bash

DESTDIR=/tmp/build_app
TEMPDIR=/tmp/build_app2
ITERATION='1'
APP_NAME='gorsync'
rm -R $DESTDIR >/dev/null 2>&1
mkdir -p $DESTDIR/usr/bin
mkdir -p $DESTDIR/usr/share/applications
rm -R $TEMPDIR >/dev/null 2>&1
mkdir -p $TEMPDIR

SAVE_DIR="${PWD}"
cd ../..
PARENT_DIR="${PWD}"
VERSION=`head -1 ./version`
./gorsync_build.sh --buildtype Release

if [ $? -eq 0 ]; then
    echo "App successfully compiled."
    cd "$SAVE_DIR"
    cp "$PARENT_DIR/$APP_NAME" "$DESTDIR/usr/bin"

    #mkdir -p $DESTDIR/etc/systemd/system
    #cp ./rpc_server.service $DESTDIR/etc/systemd/system

    cp "$PARENT_DIR/ui/gtkui/gs_schema_install.sh" "$TEMPDIR"
    cp "$PARENT_DIR/ui/gtkui/gs_schema_uninstall.sh" "$TEMPDIR"
    cp -R "$PARENT_DIR/ui/gtkui/gsettings" "$DESTDIR"
    cp ./gorsync.desktop "$DESTDIR/usr/share/applications"

    mkdir -p ./packages && cd ./packages

    packages=( "pacman" "deb" "rpm" )
    dependencies=( \
        # for Archlinux
        "--depends rsync --depends glib2 --depends gtk3 --depends libnotify" \
        # for Debian, Ubuntu
        "--depends rsync --depends libglib2.0-dev --depends libgtk-3-dev --depends libnotify-dev" \
        # for Redhat, Centos
        "--depends rsync --depends glib2-devel --depends gtk3 --depends libnotify-devel")

    for ((i=0; i<${#packages[@]};++i))
    do
        fpm -s dir -f \
            -t ${packages[i]} \
            -C $DESTDIR \
            --name $APP_NAME \
            --version $VERSION \
            --iteration $ITERATION  \
            --description "Gorsync Backup" \
            ${dependencies[i]} \
            --after-install=$TEMPDIR/gs_schema_install.sh \
            --before-remove=$TEMPDIR/gs_schema_uninstall.sh
        #    --config-files /etc
    done
    echo "done."
else
    echo "FAIL"
fi

