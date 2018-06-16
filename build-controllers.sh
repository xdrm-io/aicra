#!/bin/bash

# (1) Check argument
#--------------------------------------------------------#
if [ "$#" -lt 1 -o ! -d "$1" ]; then
	echo -e "\e[33mERROR\e[0m missing parameter\n"
	echo -e "+ expected first argument to be the controller's folder"
	exit 1;
fi



# (1) Get controller path #
BASE_DIR="`realpath $(pwd)`"
CTL_PATH="`realpath $1`";

# (2) List go files #
cd $CTL_PATH
GO_FILES="`ls *i.go **/*i.go 2>/dev/null`"

# (3) Create target directory in execution dir #
mkdir -p $BASE_DIR/controllers;

# (4) Build each file #
d=1
for file in $GO_FILES; do
	ctl="${file%.*}" # remove extension from path
	echo -e "($d) Building \e[33m$ctl\e[0m";
	# build into execution dir
	go build -buildmode=plugin -ldflags "-s -w" -o "$BASE_DIR/controllers/$ctl.so" ./$file;

	# Manage failure
	if [ "$?" -ne "0" ]; then
		echo -e "\e[31m/!\\ \e[0merror building \e[33m$ctl\e[0m\n"
	fi

	d=""`expr $d + 1`""
done;

# (5) ACK #
echo -e "\n[ \e[32mfinished\e[0m ]\n"
echo -e "files are located inside the relative \e[33m./controllers\e[0m directory"