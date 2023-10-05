#!/bin/bash

binary=./build/cosmappd
home=$HOME/.cosmappd

./build/cosmappd tx ns reserve "bob.cosmos" $(./build/cosmappd keys show alice -a --home $home --keyring-backend test) 1000uatom --from $($binary keys show bob -a --home $home --keyring-backend test)  --home $home -y