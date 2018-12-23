Gorsync Backup: GTK+ RSYNC frontend
===================================

[![Build Status](https://travis-ci.org/d2r2/go-rsync.svg?branch=master)](https://travis-ci.org/d2r2/go-rsync)
[![Go Report Card](https://goreportcard.com/badge/github.com/d2r2/go-rsync)](https://goreportcard.com/report/github.com/d2r2/go-rsync)
[![GoDoc](https://godoc.org/github.com/d2r2/go-rsync?status.svg)](https://godoc.org/github.com/d2r2/go-rsync)
[![GPLv3 License](http://img.shields.io/badge/License-GPLv3-yellow.svg)](./LICENSE)

Gorsync Backup is a best GTK+ frontend for brilliant RSYNC console utility.  Simple, but powerful.
Written completely in [Go programming language](https://golang.org/), provides responsive GUI design and intuitive interface. Might be used as training material how to write rich multi-threaded GUI application with GTK+ in Golang.



Features and benefits
----------------------

* Multiple backup profiles are supported. Moreover, each profile can be configured to get data from multiple RSYNC sources.
* 2-pass backup session approach to estimated backup volume in 1st pass. Display predicted time of completion in 2nd pass.
* Demonstrate "deduplication" on modern file systems, once previous backup sessions found (and significant time reduction in repeated backup processes). Works if backup destination is Ext3/Ext4/NTFS (employ file system hardlink feature).
* [Improved GOTK3+](https://github.com/d2r2/gotk3) library (GTK+ golang bindings) used for GUI.




Screenshots
-----------
Main form:

![image](https://raw.github.com/d2r2/go-rsync/master/docs/gorsync_main_form.png)

Preferences:

![image](https://raw.github.com/d2r2/go-rsync/master/docs/gorsync_preference_dialog.png)




Installation approaches
-----------------------

##### Build and run from sources:

* Verify, that RSYNC console utility is installed.
* Download Gorsync Backup sources (with all dependent golang libraries):
```bash
$ go get -u github.com/d2r2/go-rsync
```
* Compile and deploy application GLIB gsettings schema, with console prompt:
```bash
$ sudo ./ui/gtkui/gs_schema_install.sh
```
* Finally, run app from terminal:
```bash
$ ./gorsync_run.sh
```

##### Precompiled linux packages (deb, rpm and others) from releases:

Alternative approach to install application is to downloads installation packages from latest release, which can be found in [release page](https://github.com/d2r2/go-rsync/releases). You may find there packages for deb (Debian, Ubuntu), rpm (Fedora, Redhat) and pkg.tar.xz (Arch linux) Linux distributives.


##### Archlinux AUR repository:

One can be found in AUR repository https://aur.archlinux.org/ by name "gorsync-git" to download, compile and install latest release. On Archlinux you can use any AUR helper to install application, for instance `yaourt -S gorsync-git`.



Releases information
--------------------

##### [v0.3.1](https://github.com/d2r2/go-rsync/releases/tag/v0.3.1) (latest release):

* Internationalization implemented. Localization provided for English, Russian languages.
* Backup result's notifications: desktop notification and shell script for any kind of automation.
* Out of disk space protection for backup destination.
* A lot of improvements in algorithms and GUI.
* Significant code refactoring and rewrites.




Plans for next releases
-----------------------
Short list of preliminary anticipated features for next releases:

* More code comments and improved documentation.
* Installation packages in all major linux distribution repositories (precompiled linux packages already present).
* Application console parameters. Perhaps some CLI modes.




Gorsync Backup backup process explanation
-----------------------------------------

* As in regular [RSYNC](https://ss64.com/bash/rsync_options.html) session, Gorsync Backup is doing same job: copy files from one location (source) to another (backup destination). For instance, real life scenario would be: backing up your data from home NAS to flash hard drive attached to your notebook.

* Gorsync Backup can copy from multiple RSYNC sources at once. It could be your pictures, movies, document's from home NAS, routers and smart Linux device configuration files (/etc/... folder) and so on... Thus Gorsync Backup profile let you specify multiple separated RSYNC URL sources to get data from in single backup session and combine them all together in one destination place.

* Gorsync Backup never overwrite existing backup session destination, but use same common target root path, to put data near. For instance, your flash drive backup folder content might looks like:
```
$ <destination root folder to stores backup content>
$             ↳ ~rsync_backup_20180801-012237~
$             ↳ ~rsync_backup_20180802-013113~
$             ↳ ~rsync_backup_<date>-<time>~
...
$             ↳ ~rsync_backup_20180806-014036~
$             ↳ ~rsync_backup_(incomplete)_20180807-014024~
```
, where each specific backup session stored in separate folder with date and time in the name. "(incomplete)" phrase stands for backup, that occures at the moment. Once backup will be completed, "(incomplete)" will be removed from backup folder name. Another scenario is possible, when backup process has been interrupted for some reason: in this case "(incomplete)" phrase will never get out from folder name. But, in any case it's easy to understand where you have consistent backup results, and where not.

* In its turn, each backup folder has next regular structure:
```
$ <root folder to store backup contents on flash drive>
$             ↳ ~rsync_backup_20180801-012237~
$                       ↳ ~backup_log~.log
$                       ↳ ~backup_nodes~.signatures
$                       ↳ <folder with rsync source #1 content>
...
$                       ↳ <folder with rsync source #N content>
```
, where `~backup_log~.log` file describe all the details about the steps occurred, including info/warning/error messages if any took place. `~backup_nodes~.signatures` file contains hash imprint for all source URLs, to detect in future backup sessions same data source for "deduplication" purpose.

* Gorsync Backup is splitting backup process to the peaces. Application in every backup session is trying to find optimal data block size to backup at once. To reach this application download folders structure in 1st pass to analyze how to divide the whole process into parts. Ordinary single data block size selected to be not less than 300 MB and no more than 5 GB.



"Deduplication" capabilities
----------------------------
Once you start using Gorsync Backup on regular basis (daily/weekly/monthly), you will find soon that your backup storage will be filled with sets of almost same files (of course changes over time will provides some relatively small deviation between data sets). This redundancy might quickly exhaust your free space, but there is a real magic exists in modern file systems - "hard links"! Hard links allow do not spent space for files, which have been found in previous backup sessions unchanged. Additionally it's significantly speed up backup process. The collaboration of Gorsync Backup with RSYNC know how to activate this feature. Still you have possibility to opt out this feature in application preferences, but in general scenarios you don't need to do this.

Remember, that such legacy file systems as FAT, does not support "hard links", but successors, such as NTFS, Ext3, Ext4 and others have "hard links" supported. So, think in advance which file system to choose for your backup destination.

>*Note*: Gorsync Backup and RSYNC has some limitations with "deduplication" - they can't track file renames and relocations inside backup directory tree to save space and time in next backup session. This is not the application problem, it's RSYNC utility limitation. There is some experimental patches exist to get rid of this limitation, but not in the public RSYNC releases: you can read [this](https://lincolnloop.com/blog/detecting-file-moves-renames-rsync/) and [this](http://javier.io/blog/en/2014/08/06/rsync-rename-move.html).




Collaboration and contribution
------------------------------

If you want to contribute to the project, read next:

* Localization. Any help is appreciated to translate application to local languages. Use file ./data/assets/translate.en.toml as a source for new language translation.
* Ready to discuss proposals regarding application improvement and further development scenarios.



Contact
-------

Please use [Github issue tracker](https://github.com/d2r2/go-rsync/issues) for filing bugs or feature requests.



License
-------

Gorsync Backup is licensed under [GNU GENERAL PUBLIC LICENSE version 3](https://raw.github.com/d2r2/go-rsync/master/LICENSE) by Free Software Foundation, Inc.
