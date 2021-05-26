# *ĐArcher* Go Ethereum

This is an adaptation of official [Go Ethereum](https://github.com/ethereum/go-ethereum). 

*ĐArcher* Go Ethereum implement an blockchain environment consisting of two geth nodes to simulate blockchain reorganization.

*ĐArcher* Go Ethereum allows user to manually or programmatically control the lifecycle of each transaction. 

## Requirements

- Go: `>=1.16`
- Docker `>=20.04`
- Docker Compose `^1.29.2`

## Usage

### Build Docker Image

```bash
sh ./build-docker-image.sh
```

When building docker image, the blockchain will be initiated with the `genesis.json` and `keystore` in `blockchain` folder.

### Start Blockchain Cluster

```bash
ETHASH=$ETHASH docker-compose up
```

**Note**: replace `$ETHASH` with the folder containing the generated [`ethash`](https://eth.wiki/en/concepts/ethash/ethash).

To generate `ethash`, first build the geth:
```bash
make all
```
Then, run a script to generate geth in the `ethash` folder under the root directory of the git repository.
```bash
sh ./scripts/gen-dag.sh
```

