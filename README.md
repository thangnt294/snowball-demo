## About The Project

This is a simple demo for the snowball algorithm by Avalanche. Each transaction is simplified to an integer number, and nodes decide the validity of transactions simply by comparing the transaction to a certain threshold. If the tx number is bigger than the threshold, the transaction is valid.

## Getting started

- run the project:

  ```sh
  go run .
  ```

## Usage

Use these APIs to play around:

- POST "localhost:3000/createTx/:val": replace val with the value of the transaction you wish to create. The transaction will be created on a random node.
- GET "localhost:3000/chain": list all the local chain of all nodes.
