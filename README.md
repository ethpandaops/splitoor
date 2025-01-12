# Splitoor

- Monitor splits, validators and provide notifications on relevant events
- Deploy and manage 0xSplits v1 contracts via CLI

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

#### Get split status

Get the current controller and hash of a split.

```bash
splitoor split status \
  --el-rpc-url http://localhost:8545 \
  --contract <CONTRACT_ADDRESS> # Can omit if using mainnet/sepolia/holesky \
  --split <SPLIT_ADDRESS>
```
