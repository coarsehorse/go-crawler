#!/usr/bin/bash
# Start application in background + redirect logs to the file
nohup go run task_tracker.go >>log.log 2>&1 &
# Save last background pid
echo $! > run.pid
