#!/usr/bin/env bash
# Read this to create quota directory:
# https://www.linuxquestions.org/questions/linux-server-73/directory-quota-601140/
#
# This script help to mount destination folder to test application for "out of disk space" cases
# to imporeve recovery and error management.
#
sudo mount -o loop,rw,usrquota,grpquota /run/media/ddyakov/sda9/tmp2/limit_size.ext4 /home/ddyakov/Downloads/7777

