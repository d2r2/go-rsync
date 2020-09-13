# Find absolute path where Makedef resides with no difference how
# make utility is started (by current path, either with -C option).
mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
# Take directories part path from absolute path to Makedef file.
mkfile_dir := $(dir $(mkfile_path))
current_dir := $(notdir $(patsubst %/,%,$(dir $(mkfile_path))))
GOCMD=go
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
BINARY_NAME=gorsync
# Always reassign GOPATH here, because Makedef is used not only here, but in source archive file
# compile mode too. Move up by 4 folders from current dir to setup GOPATH as a point
# to root folder where src, bin, pkg folders are resided.
GOBUILD=eval 'GOPATH=$(mkfile_dir)../../../.. ./gorsync_build.sh --buildtype Release --output $(PWD)/$(BINARY_NAME)'
# GOBUILD=eval 'GOPATH=$(PWD) ./gorsync_build.sh --buildtype Release'

all: build

# Used for some debugging only.
print:
	$(eval GOPATH=$(shell bash -c 'echo ${PWD%/*/*/*}'))
	echo $(GOPATH)
	$(eval GOPATH=$(shell bash -c "echo ${PWD}"))
	echo $(GOPATH)
	echo ../..$(GOPATH)
	# echo '${PWD%/*/*/*}'

# Main build entry
build: 
	$(GOBUILD)

# Delete application binary from current folder and from $GOPATH/bin (if present).
clean: 
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

