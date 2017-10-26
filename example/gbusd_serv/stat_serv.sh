#!/usr/bin/env bash

echo "\n### benchmark serv info:\n---------------------------"
echo "serv-redis:\n"
cat logs/app_serv.log|grep REDISSTAT|awk -F '|' '{print $2}'|uniq -c|sort -rn|head -n 20

echo "\nserv-replication:\n"
cat logs/app_serv.log | grep STAT |  awk -F '|' '{print $2}'|uniq -c | sort -rn | head -n 20

