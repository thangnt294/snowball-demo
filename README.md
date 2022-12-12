## About The Project

This is a simple demo for the snowball algorithm by Avalanche. Each transaction is simplified to an integer number, and nodes decide the validity of transactions simply by comparing the transaction to a certain threshold. If the tx number is bigger than the threshold, the transaction is valid, and vice-versa.

The config file ava.conf is necessary for the project to run. The application will boot up ${network.size} nodes to simulate the entire network. Each node (goroutine) will occupied a port in a consecutive range, starting from ${network.startPort} upto ${network.startPort} + ${network.size}. For example, for size 10 and startPort 8000, the nodes will run within the 8000->8010 port range.

## Getting started

- run the project:

```sh
go run .
```

make sure that the config file ava.conf is available in the same directory.

## Usage

Use these APIs to play around:

- POST "localhost:${network.webPort}/createTx/:val": replace :val with the value of the transaction you wish to create. The transaction will be created on a random node in the network.
- GET "localhost:${network.webPort}/chain": list all the local chain of all nodes in the network.
- GET "localhost:${network.webPort}/neighbors/:nodeAddr": list all the neighbors of a node.
- GET "localhost:${network.webPort}/chain/:nodeAddr": list the local chain of a node.

${network.webPort} is the port configured in the config file.
