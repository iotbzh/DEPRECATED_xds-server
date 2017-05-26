#!/bin/bash

# Stop it gracefully
pkill -INT xds-server
sleep 1

# Seems no stopped, nasty kill
nbProc=$(ps -ef |grep xds-server |grep -v grep |wc -l)
if [ "$nbProc" != "0" ]; then
    pkill -KILL xds-server
fi

nbProc=$(ps -ef |grep syncthing |grep -v grep |wc -l)
if [ "$nbProc" != "0" ]; then
    pkill -KILL syncthing
    pkill -KILL syncthing-inotify
fi

