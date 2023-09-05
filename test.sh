#!/bin/bash

function calculate_port () {
  if [[ $1 -ge 0 && $1 -le 9 ]]; then
    echo "900${1}"
  elif [[ $1 -ge 10 && $1 -le 99 ]]; then
    echo "90${1}"
  elif [[ $1 -ge 100 && $1 -le 999 ]]; then
    echo "9${1}"
  fi
}

mkdir -p out
head -c 300M < /dev/urandom > out/file.txt

go run ./rpcx/smc &
sleep 0.5

(( max_client="${1}" ))
if [[ $max_client -ge 1 ]]; then
  for i in $(seq 1 "$((max_client-1))");
    do
      (( port="$(calculate_port "$i")" ))
      echo "localhost:$port"
      go run ./rpcx/store --addr "localhost:$port" --sleep 1 &
      # sleep 3
    done
fi

echo "press ctrl+c to stop the test"
(( port="$(calculate_port "$max_client")" ))
echo "localhost:$port"
go run ./rpcx/store --addr "localhost:$port" --sleep 1
pkill store
pkill smc
rm -rf out/*

echo "test finished"
