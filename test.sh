#!/bin/bash
head -c 300M < /dev/urandom > file.txt

go run ./rpcx/smc &
sleep 1

(( max_client="${1}" ))
if [[ $max_client -ge 2 ]]; then
  for i in $(seq 1 "$((max_client-1))");
    do
      if [[ $i -lt 10 ]]; then
        port="900${i}"
      elif [[ $i -lt 100 ]]; then
        port="90${i}"
      else
        port="9${i}"
      fi
      echo "localhost:$port"
      go run ./rpcx/store --addr "localhost:$port" &
      sleep 8
    done
fi

echo "press ctrl+c to stop the test"
if [[ $max_client -lt 10 ]]; then
  port="900${max_client}"
elif [[ $max_client -lt 100 ]]; then
  port="90${max_client}"
else
  port="9${max_client}"
fi
echo "localhost:$port"
go run ./rpcx/store --addr "localhost:$port"
pkill store
pkill smc
rm -f file.txt

for i in $(seq 1 "${max_client}");
  do
    if [[ $i -lt 10 ]]; then
      port="900${i}"
    elif [[ $i -lt 100 ]]; then
      port="90${i}"
    else
      port="9${i}"
    fi
    rm -f "localhost:$port.txt" &
  done

echo "test finished"