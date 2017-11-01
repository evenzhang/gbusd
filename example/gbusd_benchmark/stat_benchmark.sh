ls #!/usr/bin/env bash
echo "\n### benchmark replication info:\n---------------------------"
cat logs/app_benchmark.log.replication | grep STAT |  awk -F '|' '{print $3}'|uniq -c | sort -rn | head -n 20

echo "\n### benchmark redis info:\n---------------------------"
#cat logs/app_benchmark.log.redis |awk -F '|' '{print $3}'|uniq -c|sort -rn| head -n 20
cat logs/app_benchmark.log.redis |grep REDISSTAT|awk -F '|' '{print $2}'|uniq -c|sort -rn| head -n 20

echo "\n### benchmark serv info:\n---------------------------"
echo "serv-redis:\n"
cat logs/app_benchmark.log.serv|grep REDISSTAT|awk -F '|' '{print $2}'|uniq -c|sort -rn| head -n 20

echo "\nserv-replication:\n"
cat logs/app_benchmark.log.serv| grep DBSTAT |  awk -F '|' '{print $3}'|uniq -c | sort -rn | head -n 20
