# Fee-station

## Design

stationd: api server

syncerd: sync txs when user transfer to pool address and update prices of token

payerd: send fis to user

## How to use

```sh
make build
# after config conf_station.toml
./build/stationd -C ./conf_station.toml
# after config conf_payer.toml
./build/payerd -C ./conf_payer.toml
# after config conf_syncer.toml
./build/syncerd -C ./conf_syncer.toml
```