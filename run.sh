#!/usr/bin/bash
# Start application in background + redirect logs to the file
go build task_tracker.go
nohup ./task_tracker >>log.log 2>&1 &
# Save last background pid
echo $! > pid.pid
