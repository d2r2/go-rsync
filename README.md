Gorsync Backup: GTK+ RSYNC frontend
===================================

[![Build Status](https://travis-ci.org/d2r2/go-rsync.svg?branch=master)](https://travis-ci.org/d2r2/go-rsync)
[![Go Report Card](https://goreportcard.com/badge/github.com/d2r2/go-rsync)](https://goreportcard.com/report/github.com/d2r2/go-rsync)
[![GoDoc](https://godoc.org/github.com/d2r2/go-rsync?status.svg)](https://godoc.org/github.com/d2r2/go-rsync)
[![GPLv3 License](http://img.shields.io/badge/License-GPLv3-yellow.svg)](./LICENSE)

ANNOUNCEMENT: Translation to local languages wanted!!! Please read [collaboration and contribution](#collaboration-and-contribution).
----------

About
------------

Gorsync Backup is a GTK+ frontend for brilliant [RSYNC](https://download.samba.org/pub/rsync/rsync.html) console utility.  Written completely in [Go programming language](https://golang.org/), provides responsive GUI design and intuitive interface.

Gorsync Backup is not an all-purpose RSYNC wrapper. It was created to keep in mind how to do regular backup of your personal data from home NAS (either small business file server with RSYNC service enabled) with minimal effort. Gorsync Backup doesn't implement user interface for any possible RSYNC option and case, but use all the best from powerful RSYNC utility to perform regular data backup easily and quickly.

If you do not want to pay monthly fee for cloud space to backup your personal data to the cloud, but instead prefer to buy inexpensive mass-market 2-4 TB portable external hard drive, then your average backup scheme will looks like: attach USB hard drive to workstation/notebook with desktop Linux OS, start Gorsync Backup application, choose backup profile (once configured), run the backup process and you can go and enjoy your coffee cup.

The application will do the rest: estimate size of data to backup, display extensive progress information and once it done report about completion. If this is not your first backup, then application will significantly speed up backup process, because it keeps previous backup history in destination location and has "deduplication" capabilities, which allows the application do not copy over the network data that was backed up in previous sessions (withal it saves a lot of space): in my experience repeated backup of whole photography archive which exceed 650 GB from my home NAS with new pictures from last few months, takes 15 minutes.


Features and benefits
----------------------

* Multiple backup profiles are supported. Moreover, each profile can be configured to get data from multiple RSYNC sources/modules. Password protected RSYNC modules supported.
* 2-pass backup session approach to estimated backup volume in 1st pass. Display predicted time of completion in 2nd pass.
* Demonstrate "deduplication" on modern file systems, once previous backup sessions found (and significant time reduction in repeated backup processes). Works if backup destination is ext3/ext4/NTFS (employ file system hardlink feature).
* [Improved GOTK3+](https://github.com/d2r2/gotk3) library (GTK+ golang bindings) used for GUI.



Screenshots
-----------
Main form:

![image](https://raw.github.com/d2r2/go-rsync/master/docs/gorsync_main_form.png)

Preferences:

![image](https://raw.github.com/d2r2/go-rsync/master/docs/gorsync_preference_dialog.png)




Installation approaches
-----------------------

#### Build and run from sources.

Compilation from source successfully tested with Linux and FreeBSD systems.

Steps to build and run:

* Verify that GLIB2, GTK3, libnotify libraries, RSYNC console utility, go/gcc/clang compilers and other development tools are present in the system.
* Download Gorsync Backup sources (with all dependent golang libraries):
```bash
$ go get -u github.com/d2r2/go-rsync
```
* Compile and deploy application GLIB gsettings schema, with console prompt:
```bash
$ cd ./ui/gtkui/
$ sudo ./gs_schema_install.sh
```
* Finally, run app from terminal:
```bash
$ ./gorsync_run.sh
```
, either compile application binary:
```bash
$ ./gorsync_build.sh --buildtype Release|Development
```


#### Precompiled linux packages (deb, rpm and others) from releases.

Alternative approach to install application is to downloads installation packages from latest release, which can be found in [release page](https://github.com/d2r2/go-rsync/releases). You may find there packages for deb (Debian, Ubuntu), rpm (Fedora, Centos, Redhat) and pkg.tar.xz (Arch linux) Linux distributives.


#### Archlinux AUR repository.

One can be found in AUR repository https://aur.archlinux.org/ by name "gorsync-git" to download, compile and install latest release. On Archlinux you can use any AUR helper to install application, for instance `yaourt -S gorsync-git`.



Releases information
--------------------

#### [v0.3.3](https://github.com/d2r2/go-rsync/releases/tag/v0.3.3) (latest release):

* RSYNC transfer options override implemented in profile settings.
* Lot of small UI improvements.
* Replace UI animation and customization implemented via imperative programming, with GTK+ CSS declarative application styling.
* Both Linux and FreeBSD supported.
* Adaptation to latest GLIB 2.62, GTK+ 3.24.
* Updated documentation and help.
* Bugs fixed.


#### [v0.3.2](https://github.com/d2r2/go-rsync/releases/tag/v0.3.2):

* Password protected RSYNC module supported.
* Option to change files permission during RSYNC backup process.
* First attempt to create project site: [https://gorsync.github.io/](https://gorsync.github.io/) on the base of superior [Beautiful Jekyll](https://deanattali.com/beautiful-jekyll/) template.
* Adaptation to latest GLIB 2.60, GTK+ 3.24.
* Updated documentation and help.
* Bugs fixed.


#### [v0.3.1](https://github.com/d2r2/go-rsync/releases/tag/v0.3.1):

* Internationalization implemented. Localization provided for English, Russian languages.
* Backup result's notifications: desktop notification and shell script for any kind of automation.
* Out of disk space protection for backup destination.
* A lot of improvements in algorithms and GUI.
* Significant code refactoring and rewrites.




Plans for future releases
-------------------------

Short list of preliminary anticipated features for next releases:

* More code comments and improved documentation.
* Installation packages for all major Linux distributives.




Gorsync Backup backup process description
-----------------------------------------

As in regular [RSYNC](https://ss64.com/bash/rsync_options.html) session, Gorsync Backup is doing same job: copy files from one location (source) to another (backup destination).

Gorsync Backup might be configured to copy from multiple RSYNC sources. It could be your pictures, movies, document's from home NAS, routers and smart Linux device configuration files (/etc/... folder) and so on... Thus Gorsync Backup profile let you specify multiple separated RSYNC URL sources to get data from in single backup session and combine them all together in one destination place.

Gorsync Backup never overwrite previous backup session data, but use same common target root path, to put data near by in new folder. For instance, your flash drive backup folder content might looks like:
```
$ <destination root folder to stores backup content>
$             ↳ ~rsync_backup_20180801-012237~
$             ↳ ~rsync_backup_20180802-013113~
$             ↳ ~rsync_backup_<date>-<time>~
...
$             ↳ ~rsync_backup_20180806-014036~
$             ↳ ~rsync_backup_(incomplete)_20180807-014024~
```
, where each specific backup session stored in separate unique folder with date and time in the name. "(incomplete)" phrase stands for backup, that occurs at the moment. When backup is completed, "(incomplete)" will be removed from backup folder name. Another scenario is possible, when backup process has been interrupted for some reason: in this case "(incomplete)" phrase will never get out from folder name. But, in any case it's easy to understand where you have consistent backup results, and where not.

In its turn, each backup folder has next regular structure:
```
$ <destination root folder to stores backup content>
$             ↳ ~rsync_backup_20180801-012237~
$                       ↳ ~backup_log~.log
$                       ↳ ~backup_nodes~.signatures
$                       ↳ <folder with rsync source #1 content>
...
$                       ↳ <folder with rsync source #N content>
```
, where `~backup_log~.log` file describe all the details about the steps occurred, including info/warning/error messages if any took place. `~backup_nodes~.signatures` file contains hash imprint for all source URLs, to detect in future backup sessions same data sets for "deduplication" activation (so, never delete this file from backup history, otherwise "deduplication" will not work for future backup sessions).

Gorsync Backup gives extra flexibility in copying data: you can configure application to skip copying some data from single RSYNC source. For this, you should place empty file in the folder with specific name `!!!__SKIPBACKUP__!!!` (name can be changed in preference), which instruct application not to copy specific folder content (including subfolders).

Gorsync Backup is splitting backup process to the peaces. The application in every backup session is trying to find optimal data block size to backup at once. To reach this application download folders structure in 1st pass to analyze how to divide the whole process into parts. Ordinary single data block size selected to be not less than 300 MB and not exceed 5 GB.



Regular backup and "deduplication" capabilities
-----------------------------------------------
Once you start using Gorsync Backup on regular basis (monthly/weekly/daily), you will find soon that your backup storage will be filled with set of almost same files (of course changes over time will provides some relatively small deviation between data sets). This redundancy might quickly exhaust your free space, but there is a real magic exists in modern file systems - "hard links"! Hard links allow the application not to spend space for the files, which have been found in previous backup sessions unchanged. Additionally it's significantly speed up backup process. The collaboration of Gorsync Backup application with RSYNC utility know how to activate this feature. Still you have possibility to opt out this feature in application preferences, but in general scenario you don't need to do this.

Be aware, that such legacy file systems as FAT, does not support "hard links", but successors, such as NTFS, ext3, ext4 and others have "hard links" supported (via low level file system "inode" concept). Hence no deduplication is possible for FAT. So, think in advance which file system to choose for your backup destination.

>*Note*: Gorsync Backup and RSYNC on its own has some limitations with "deduplication" - they can't track file renames and relocations inside backup directory tree to save space and time in the following backup sessions. This is not the application problem, it's RSYNC utility limitation. There is some experimental patches exist to get rid of this limitation, but not in the public RSYNC releases, which the application fully relies on: you can read [this](https://lincolnloop.com/blog/detecting-file-moves-renames-rsync/) and [this](http://javier.io/blog/en/2014/08/06/rsync-rename-move.html).

>*Note*: Remember, that changing RSYNC transfer options between backup sessions may affect deduplication behavior for same backup data set. For instance, just enabled "transfer file permissions" might eliminate deduplication, even when backup data from previous backup sessions are present (but was made with disabled "transfer file permissions" option). Using NTFS file system in destination and having "transfer file permissions" option enabled, may cause that deduplication will not work in some cases.



Collaboration and contribution
------------------------------

If you want to contribute to the project, please, read:

* User interface translation and localization required. Any help is appreciated to translate application to local languages (so far only English and Russian are supported).
Use `./data/assets/translate.en.toml` file as an English language source for new language translation. You should modify each section/paragraph value in the file written in [TOML](https://en.wikipedia.org/wiki/TOML) format, for instance:
```
[AboutDlgDoNotShowCaption]
other = "Do not show this information at application startup"

[PrefDlgProfileNameExistsWarning]
other = "Profile with name \"{{.ProfileName}}\" already exists. Please, correct the name"
```
, to get in the end translated version (do not translate identifiers in double brackets {{...}}). Finally, change "en" suffix in the file name to corresponding language code, and propose translation via project pull request, either mail file directly to denis.dyakov@gmail.com. If you still have any question, please, let me know.
* Ready to discuss proposals regarding application improvement and further development.



Contact
-------

Please use [Github issue tracker](https://github.com/d2r2/go-rsync/issues) for filing bugs or feature requests.



License
-------

Gorsync Backup is licensed under [GNU GENERAL PUBLIC LICENSE version 3](https://raw.github.com/d2r2/go-rsync/master/LICENSE) by Free Software Foundation, Inc.
