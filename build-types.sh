#!/bin/bash

# (1) Get default path #
DEFAULT_PATH="`realpath $GOPATH`/src/git.xdrm.io/go/aicra/checker/default";
# fail if do not exist
test -d $DEFAULT_PATH || (echo "$DEFAULT_PATH does not exists.. aborting"; exit 1) || exit 1;

# (2) List go files #
GO_FILES="`ls $DEFAULT_PATH/**/*.go`"

# (3) Create target directory #
mkdir -p ./types;

# (4) Build each file #
d=1
for Type in $GO_FILES; do
	plugin_name="`basename $(dirname $Type)`"
	echo -e "($d) Building \e[33m$plugin_name\e[0m";
	go build -buildmode=plugin -o "./types/$plugin_name.so" $Type;

	# Manage failure
	if [ "$?" -ne "0" ]; then
		echo -e "\e[31m/!\\ \e[0merror building \e[33m$plugin_name\e[0m\n"
	fi

	d=""`expr $d + 1`""
done;

# (5) ACK #
echo -e "\n[ \e[32mfinished\e[0m ]\n"
echo -e "files are located inside the relative \e[33m./types\e[0m directory"