#!/bin/sh

## Define from ENV file like this
# PROJ_NAME="someapp"
# LOCAL_DIR=$(pwd)"/"
# REMOTE_DIR="/funny/"$PROJ_NAME"/approot"
# REMOTE_USER="xxxx"
# REMOTE_PORT="xxxx"
# REMOTE_HOST="192.168.xxx.xxx"
. $(pwd)"/deploy.conf"

## Prepare configs directory


rsync -e "ssh -p $REMOTE_PORT" \
	--exclude=".gitignore" \
	--exclude="uploaded" \
	--exclude=".env*"	\
	--exclude="logs/" \
	--exclude=".git" \
	-PLSluvr --del --no-perms --no-t \
	$LOCAL_DIR $REMOTE_USER@$REMOTE_HOST:$REMOTE_DIR

## clean

echo "Transfered to webserver host $REMOTE_HOST:$REMOTE_PORT: Ok!"
exit 100
