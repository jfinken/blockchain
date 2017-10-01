blockchain
---

#### Overview

A blockchain, hosted over HTTP, implemented in Go.

#### Installation

#### Usage

    go build; ./blockchain

Then

    # Observe the current state of the blockchain
    curl http://localhost:8181/chain

    # Register another node to the network (for consensus examples)
    curl -H "Content-Type: application/json" -X POST -d '{"ip": "10.0.0.87"}' http://localhost:8181/nodes/register

    # POST a new transaction bound for a future block
    curl -H "Content-Type: application/json" -X POST -d '{"sender":"foo","receiver":"99d1d48c-b112-498b-a444-d5c0277cac2d", "amount":50}' http://localhost:8181/transactions/new

    # Mine and forge a new block and be rewarded with coin.
    curl http://localhost:8181/mine