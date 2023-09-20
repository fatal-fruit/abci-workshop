# ABCI++ Workshop

Joint workshop between [Cosmos SDK](https://docs.cosmos.network/main) and [CometBFT](https://cometbft.com/) for [Hackmos](https://cosmoverse.org/hackmos) at [Cosmoverse](https://cosmoverse.org/).

Supported by [Informal Systems](https://informal.systems/) and [Binary Builders](https://binary.builders/).

## Concepts
**Overview**

The goal of this workshop is to demonstrate to developers how they might think about utilizing ABCI++ and new features in the Eden release of the Cosmos SDK.

In part 1, we explore the application and discuss a potential vector for MEV by leveraging custom block building in `PreparePropsoal` for a chain running perpetual nameservice auctions.

In part 2 and 3, we build a solution to mitigate auction front running by extending functionality in `ExtendVote`, `PrepareProposal`, and `ProcessProposal`. This solution assumes an honest 2/3 majority of validators are not colluding to front run transactions.

At **H-1** during `ExtendVote`, we check the mempool for unconfirmed transactions and select all auction bids. Validators submit their Vote Extension with a list of all bids available in their mempool.
Additionally we implement a custom app side `ThresholdMempool`, which guarantees that transactions can only be included in a proposal if they have been seen by `ExtendVote` at H-1.

At **H** during `PrepareProposal`, the validator will process all bids included in Vote Extensions from H-1. It will inject this result into a Special Transaction to be included in the proposal.
During the subsequent `ProcessProposal`, validators will check if there are any bid transactions. Bids included in the proposal will be validated against the bids included in the Special Transaction.
If a bid included in the proposal does not meet the minimum threshold of inclusion frequency in Vote Extensions from H-1, the proposal is rejected.

![](./figures/diagram.png)

## Content

### Part 1
1. Getting Started
2. Exploring MEV Mitigation

### Part 2
1. Submit Proof of Existence with Vote Extensions
2. Winning the Race: Modifying the Mempool

### Part 3
1. Collating Evidence in Prepare Proposal
2. Detecting & Deflecting Misbehavior with Process Proposal

<hr>

## Developing

### Setup

**Dependencies**
- [Go 1.21](https://go.dev/dl/)
- [Jq](https://jqlang.github.io/jq/)

#### Start A Single Chain
<<<<<<< Updated upstream
> Note: Running the provider on a single chain will affect liveness. This setup is best used to demonstrate Part 1.
=======
> Note: Running the provider on a single chain will affect liveness. This setup is best used to demonstrate Part 1. 
>>>>>>> Stashed changes
> First checkout [part-1-2](https://github.com/fatal-fruit/abci-workshop/tree/glnro/part-1-2) and navigate to the [single node demo](#single-node).

```
make start-localnet

jq '.consensus.params.abci.vote_extensions_enable_height = "2"' ~/.cosmappd/config/genesis.json > output.json && mv output.json ~/.cosmappd/config/genesis.json

./build/cosmappd start --val-key val1 --run-provider false
```

#### Start a 3 Validator Network

Make sure to run `make build`

```shell
./scripts/configure.sh
```

Read Logs
```shell

tail -f $HOME/cosmos/nodes/beacon/logs
tail -f $HOME/cosmos/nodes/val1/logs
tail -f $HOME/cosmos/nodes/val1/logs
```

Query a node
```shell
./scripts/query-beacon-status.sh
```

List Available User Keys
```shell
./scripts/list-beacon-keys.sh
```

### Demo

> **Vote Extensions** are enabled from Height 2, so make sure not to submit transactions until H+1 has been comitted.

#### 3 Validator Network
In the 3 validator network, the Beacon validator has a custom transaction provider enabled.
It might take a few tries before the transaction is picked up and front ran by the Beacon.

After submitting the following transaction, we should be able to see the proposal accepted or rejected in the logs.
Note that it is best to submit the transaction after the Beacon has just proposed a successful proposal.
```shell
./scripts/reserve.sh "bob.cosmos"
```
Query to verify the name has been reserved
```shell
./scripts/whois.sh "bob.cosmos"
```
If the Beacon attempts to front run the bid, we will see the following logs during `ProcessProposal`
```shell
2:47PM ERR ❌️:: Detected invalid proposal bid :: name:"bob.cosmos" resolveAddress:"cosmos1wmuwv38pdur63zw04t0c78r2a8dyt08hf9tpvd" owner:"cosmos1wmuwv38pdur63zw04t0c78r2a8dyt08hf9tpvd" amount:<denom:"uatom" amount:"2000" >  module=server
2:47PM ERR ❌️:: Unable to validate bids in Process Proposal :: <nil> module=server
2:47PM ERR prevote step: state machine rejected a proposed block; this should not happen:the proposer may be misbehaving; prevoting nil err=null height=142 module=consensus round=0
```

#### Single Node

```shell
./build/cosmappd tx ns reserve "bob.cosmos" $(./build/cosmappd keys show alice -a --keyring-backend test) 1000uatom --from $(./build/cosmappd keys show bob -a --keyring-backend test) -y
```
Query to verify the name has been reserved.

If running the provider on the `part-1-2` branch, we can see that the owner and resolve address are different than the original transaction we submitted.
```shell
./build/cosmappd q ns whois "bob.cosmos" -o json

./build/cosmappd keys list --keyring-backend test --output json | jq
```

## Resources

Official Docs
- [CometBFT](https://docs.cometbft.com/v0.37/)
- [Cosmos SDK](https://docs.cosmos.network/main)

ABCI++
- [ACBI++ Spec: Basic Concepts](https://github.com/cometbft/cometbft/blob/main/spec/abci/abci++_basic_concepts.md#consensusblock-execution-methods)
- [ABCI++ Spec: Application Requirements](https://github.com/cometbft/cometbft/blob/main/spec/abci/abci%2B%2B_app_requirements.md)
- [Skip's POB Article](https://ideas.skip.money/t/x-builder-the-first-sovereign-mev-module-for-protocol-owned-building/57)
- Videos
  - [Sergio from Informal: Presentation on ABCI++](https://youtube.com/watch?v=cAR57hZaJtM)
  - [Evan from Celestia: Possible Applications of ABCI++](https://www.youtube.com/watch?v=VGdIZLVYoRs)

Building Applications
- [Facu from Binary Builders: Building Cosmos SDK Modules](https://www.youtube.com/watch?v=9kK9uzwEeOE)
- [Interchain Developer Academy Playlist](https://www.youtube.com/watch?v=1_ottIKPfI4&list=PLE4J1RDdNh6sTSDLehUpp7vqvm2WuFWNU&pp=iAQB)
- [Cosmod's Basic module template](https://github.com/cosmosregistry/example)