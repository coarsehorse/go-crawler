#!/usr/bin/bash
lastPid=$(cat run.pid)
if [ -z "$lastPid" ]
then
    echo "\$lastPid is empty. Skipping stopping step"
else
    kill -9 $lastPid;
