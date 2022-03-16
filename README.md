# Fee-station

## Design

stationd: api server

payerd: sync txs when user transfer to pool address, update prices of token, send fis to user

## How to use

```sh
make build
# after config conf_station.toml
./build/stationd -C ./conf_station.toml
# after config conf_payer.toml
./build/payerd -C ./conf_payer.toml
```