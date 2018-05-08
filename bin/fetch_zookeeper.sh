#!/bin/bash

ZK="zookeeper-3.4.9"

# Fetch
wget https://archive.apache.org/dist/zookeeper/${ZK}/${ZK}.tar.gz

# Extract
tar -zxf ${ZK}.tar.gz

# Trim
rm ${ZK}.tar.gz
find ${ZK} -type f ! -iname "${ZK}-fatjar.jar" -delete
find ${ZK} -depth -type d -exec rmdir {} \;
