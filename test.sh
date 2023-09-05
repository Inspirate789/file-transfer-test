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
if [[ $max_client -ge 2 ]]; then
  for i in $(seq 0 "$((max_client-2))");
    do
      (( port1="$(calculate_port "$((2*i))")" ))
      (( port2="$(calculate_port "$((2*i+1))")" ))
      echo "localhost:$port1"
      echo "localhost:$port2"
      go run ./rpcx/store --addr_store "localhost:$port1" --addr_file_service "localhost:$port2" &
      # sleep 3
    done
fi

echo "press ctrl+c to stop the test"
(( port1="$(calculate_port "$((2*(max_client-1)))")" ))
(( port2="$(calculate_port "$((2*max_client-1))")" ))
echo "localhost:$port1"
echo "localhost:$port2"
go run ./rpcx/store --addr_store "localhost:$port1" --addr_file_service "localhost:$port2"
pkill store
pkill smc
rm -rf out/*

echo "test finished"