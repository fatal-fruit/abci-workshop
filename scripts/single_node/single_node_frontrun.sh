binary=./build/cosmappd
home=$HOME/.cosmappd

$binary tx ns reserve "bob.cosmos" $($binary keys show alice -a --home $home --keyring-backend test) 1000uatom --from $($binary keys show bob -a --home $home --keyring-backend test)  --home $home -y