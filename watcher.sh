#!/bin/bash
# self WATCH_SRCS YARN LOG_LEVEL MODE
if [ -n "$1" ]; then
    MODE=$1
fi

if [ -n "$2" ]; then
    export WATCH_SRCS=$2
fi

if [ -n "$3" ]; then
    export YARN=$3
fi

if [ -n "$4" ]; then
    export LOG_LEVEL=$4
fi

if [ -n "$5" ]; then
    export OTHER_ARGS=$5
fi


cleanup(){
	echo "##killing server##"
	ps aux |grep -v grep |grep 'make run-server' |awk '{print $2}' |xargs kill
	exit 0
}

build-server(){
	while true; do
		make run-server YARN="$YARN" LOG_LEVEL="$LOG_LEVEL" OTHER_ARGS="$OTHER_ARGS"&
		inotifywait -e close_write,moved_to,create $WATCH_SRCS
		ps aux|grep -v grep|grep "make run-server"|awk '{print $2}' |xargs kill
	done

}

dev-server(){
while true; do
		make run-server-nosync YARN="$YARN" LOG_LEVEL="$LOG_LEVEL" OTHER_ARGS="$OTHER_ARGS"&
		inotifywait -e close_write,moved_to,create $WATCH_SRCS
		ps aux|grep -v grep|grep "make run-server"|awk '{print $2}' |xargs kill
	done

}

dev-client(){
	while true; do
		make run-client-nosync YARN="$YARN" LOG_LEVEL="$LOG_LEVEL" OTHER_ARGS="$OTHER_ARGS"&
		inotifywait -e close_write,moved_to,create $WATCH_SRCS
		sleep $((RANDOM % 4)) # avoid compile-race
		ps aux|grep -v grep|grep "make run-client"|awk '{print $2}' |xargs kill
	done

}
dev(){
	echo "##replacing config.js API_PREFIX##"
	sed -i 's|"/api"|"//localhost:4533/api"|g' ./frontend-repo/src/assets/config.js
	echo "oping multi-windows"
	tmux new-session -n watcher "bash -c 'source ./watcher.sh&& dev-server'" \
		\; split-window -h "make dev-frontend YARN='$YARN'" \
		\; split-window -h "bash -c 'source ./watcher.sh&& dev-client'" \
		\; select-layout even-horizontal
	echo "##resume config.js API_PREFIX##"
	sed -i 's|"//localhost:4533/api"|"/api"|g' ./frontend-repo/src/assets/config.js

}


if [[ -n "$MODE" ]];then
	trap cleanup INT 
	echo "##please exit this script by Ctrl-C!##"
	echo "##It is better to close autosave##"
	if [[ "$MODE" == "build" ]]; then
			build-server 
	elif [[ "$MODE" == "dev" ]]; then
			dev
	fi
fi