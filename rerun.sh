#!/usr/bin/bash
kill -9 $(cat run.pid);
chmod +x run.sh;
./run.sh
