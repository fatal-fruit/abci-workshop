#!/bin/bash

alice=$(cosmappd keys show alice -a)
bob=$(cosmappd keys show bob -a)
beatrice=$(cosmappd keys show beatrice -a)






cosmappd tx bank send bob $($BINARY keys show alice -a --home $HOME/cosmos/nodes/beacon --keyring-backend test)  100uatom -y --output json
cosmappd tx bank send alice$($BINARY keys show bob -a --home $HOME/cosmos/nodes/beacon --keyring-backend test)  10uatom -y --output json
cosmappd tx bank send bob $($BINARY keys show alice -a --home $HOME/cosmos/nodes/beacon --keyring-backend test)  100uatom -y --output json
cosmappd tx bank send bob $($BINARY keys show alice -a --home $HOME/cosmos/nodes/beacon --keyring-backend test)  100uatom -y --output json