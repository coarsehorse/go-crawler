#!/usr/bin/bash
lastPid=$(cat pid.pid)
if [ -z "$lastPid" ]
then
    echo "\$lastPid is empty. Skipping stopping step"
else
    echo "Last application instance(pid:$lastPid) has been stopped"
    kill -9 $lastPid
fi
