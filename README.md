# Splitoor

- Monitor splits, validators and provide notifications on relevant events
- Deploy and manage 0xSplits v1 contracts via CLI


## Getting Started

### Download a release

Download the latest release from the [Releases page](https://github.com/ethpandaops/splitoor/releases). Extract and run with:

```bash
./splitoor --help
```

### Docker

Available as a docker image at [ethpandaops/splitoor](https://hub.docker.com/r/ethpandaops/splitoor/tags)
#### Images

- `latest` - distroless, multiarch
- `latest-debian` - debian, multiarch
- `$version` - distroless, multiarch, pinned to a release (i.e. `0.1.0`)
- `$version-debian` - debian, multiarch, pinned to a release (i.e. `0.1.0-debian`)

**Quick start**

```bash
docker run -d  --name splitoor -v $HOST_DIR_CHANGE_ME/config.yaml:/opt/splitoor/config.yaml -p 9090:9090 -it ethpandaops/splitoor:latest monitor --config /opt/splitoor/config.yaml;
docker logs -f splitoor;
```

### Kubernetes via Helm

- [splitoor](https://github.com/skylenet/ethereum-helm-charts/tree/master/charts/splitoor)

```bash
helm repo add ethereum-helm-charts https://ethpandaops.github.io/ethereum-helm-charts

# monitor
helm install splitoor ethereum-helm-charts/splitoor -f your_values.yaml
```

### Building yourself

1. Clone the repo
   ```sh
   go get github.com/ethpandaops/splitoor
   ```
1. Change directories
   ```sh
   cd ./splitoor
   ```
1. Build the binary
   ```sh  
    go build -o splitoor .
   ```
1. Run the monitor
   ```sh  
    ./splitoor monitor --config example_server_config.yaml
   ```

## Usage

Splitoor has two main running modes: a monitor and interacting with 0xSplits v1 contracts.

### Monitor

Look at the [example_config.yaml](./example_config.yaml) for an example configuration.

```bash
# defaults to config.yaml in current directory
splitoor monitor --config <CONFIG_FILE>
```

### 0xSplits

Splitoor can interact with [0xSplits v1](https://docs.splits.org/core/split) contracts.

#### Deploy contract

Deploy the 0xSplits v1 contract

> Not required if using mainnet/sepolia/holesky as contracts are already deployed

```bash
splitoor deploy-contract \
  --el-rpc-url http://localhost:8545 \
  --deployer-address <DEPLOYER_ADDRESS> \
  --deployer-private-key <DEPLOYER_PRIVATE_KEY>
```

#### Create split

Create a new split.

```bash
splitoor split create \
  --el-rpc-url http://localhost:8545 \
  --deployer-address <DEPLOYER_ADDRESS> # does not need to be the controller or a recipient \
  --deployer-private-key <DEPLOYER_PRIVATE_KEY> \
  --contract <CONTRACT_ADDRESS> # Can omit if using mainnet/sepolia/holesky \
  --controller <CONTROLLER_ADDRESS> \
  --recipients <RECIPIENT_1_ADDRESS>,<RECIPIENT_2_ADDRESS> \
  --percentages 500000,500000 # 50%, 50%
```

#### Update split

Update a split's recipients and percentages.

```bash
splitoor split update \
  --el-rpc-url http://localhost:8545 \
  --deployer-address <DEPLOYER_ADDRESS> # must be the controller of the split \
  --deployer-private-key <DEPLOYER_PRIVATE_KEY> \
  --contract <CONTRACT_ADDRESS> # Can omit if using mainnet/sepolia/holesky \
  --split <SPLIT_ADDRESS> \
  --recipients <RECIPIENT_1_ADDRESS>,<RECIPIENT_2_ADDRESS> \
  --percentages 600000,400000 # 60%, 40%
```

#### Calculate split hash

Calculate the keccak256 hash of split recipients and percentages.

```bash
splitoor split calculate-hash \
  --recipients <RECIPIENT_1_ADDRESS>,<RECIPIENT_2_ADDRESS> \
  --percentages 600000,400000 # 60%, 40%
```

#### Get split status

Get the current controller and hash of a split.

```bash
splitoor split status \
  --el-rpc-url http://localhost:8545 \
  --contract <CONTRACT_ADDRESS> # Can omit if using mainnet/sepolia/holesky \
  --split <SPLIT_ADDRESS>
```

#### Distribute ETH

Distribute ETH to the split's recipients.

```bash
splitoor split distribute \
  --el-rpc-url http://localhost:8545 \
  --deployer-address <DEPLOYER_ADDRESS> \
  --deployer-private-key <DEPLOYER_PRIVATE_KEY> \
  --contract <CONTRACT_ADDRESS> # Can omit if using mainnet/sepolia/holesky \
  --split <SPLIT_ADDRESS> \
  --recipients <RECIPIENT_1_ADDRESS>,<RECIPIENT_2_ADDRESS> \
  --percentages 600000,400000 # 60%, 40%
```

#### Get ETH balance

Get the distributed ETH balance of an address on the splits contract. This can be from multiple splits.

```bash
splitoor split balance \
  --el-rpc-url http://localhost:8545 \
  --contract <CONTRACT_ADDRESS> # Can omit if using mainnet/sepolia/holesky \
  --address <ADDRESS>
```

#### Withdraw

Withdraw from the splits contract.

```bash
splitoor split withdraw \
  --el-rpc-url http://localhost:8545 \
  --private-key <DEPLOYER_PRIVATE_KEY> \
  --contract <CONTRACT_ADDRESS> # Can omit if using mainnet/sepolia/holesky \
  --address <ADDRESS>
  --withdraw-eth # omit if only withdrawing ERC20s \
  # --tokens <TOKEN_1_ADDRESS>,<TOKEN_2_ADDRESS> # optional, comma separated list of tokens addresses to withdraw
```
