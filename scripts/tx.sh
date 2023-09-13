#!/bin/bash

alice=$(cosmappd keys show alice --address)
bob=$(cosmappd keys show bob --address)
beatrice=$(cosmappd keys show beatrice --address)

get_name () {
    for people in [$alice,$bob,$beatrice]; do
        stringify_name $1
    done;
}

stringify_name () {
    case $1 in 
        $alice)
            echo "alice"
            ;;
        $bob)
            echo "bob"
            ;;
        $beatrice)
            echo "beatrice"
            ;;
         *)
            echo $1
            ;;
    esac
}

# Normally mempool were FIFO, so it the block should have been by time of transaction receipt
# In this case BEATRICE - ALICE - BOB (try this with `cosmappd start --mempool-type none`)
# However, with the fee mempool, the transactions or ordered by fees, so in the block it will be ordered as
# BOB - ALICE - BEATRICE (try this with `cosmappd start --mempool-type fee`)
echo "--> sending transactions in the order beatrice, alice, bob"
cosmappd tx bank send beatrice $alice 10uatom -y --output json > /dev/null
tx=$(cosmappd tx bank send alice $bob 10uatom --fees 10uatom -y --output json | jq -r .txhash)
cosmappd tx bank send bob $beatrice 10uatom --fees 100uatom -y --output json > /dev/null

echo "--> sleeping the block time timeout duration"
sleep 15

# query which block those txs have been included into
height=$(cosmappd q tx $tx --type hash --output json | jq .height -r)

# get all txs
txs=$(cosmappd q block $height | jq .block.data.txs -r | jq -c '.[]')

echo "--> printing transaction order in block $height"

for rawTx in $txs; do
    get_name $(cosmappd tx decode $(echo $rawTx | jq -r) | jq -r .body.messages[0].from_address)
done;