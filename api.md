# API doc


## 0. status code

```go
	codeSuccess               = "80000"
	codeParamParseErr         = "80001"
	codeSymbolErr             = "80002"
	codeStafiAddressErr       = "80003"
	codeTxHashErr             = "80004"
	codePubkeyErr             = "80005"
	codeInternalErr           = "80006"
	codePoolAddressErr        = "80007"
	codeTxDuplicateErr        = "80008"
	codeTokenPriceErr         = "80009"
	codeInAmountFormatErr     = "80010"
	codeMinOutAmountFormatErr = "80011"
	codePriceSlideErr         = "80012"
	codeMinLimitErr           = "80013"
	codeMaxLimitErr           = "80014"
	codeSwapInfoNotExistErr   = "80015"
```

## 1. get pool info

### (1) description

*  get pool info

### (2) path

* /feeStation/api/v1/station/poolInfo

### (3) request method

* get

### (4) request payload 

* null
 
### (5) response
* include status、data、message fields
* status、message must be string format,data must be object

| grade 1 | grade 2      | grade 3     | type   | must exist? | encode type | description                                      |
| :------ | :----------- | :---------- | :----- | :---------- | :---------- | :----------------------------------------------- |
| status  | N/A          | N/A         | string | Yes         | null        | status code                                      |
| message | N/A          | N/A         | string | Yes         | null        | status info                                      |
| data    | N/A          | N/A         | object | Yes         | null        | data                                             |
|         | poolInfoList | N/A         | list   | Yes         | null        | list                                             |
|         |              | symbol      | string | Yes         | null        | native token like `uatom`                        |
|         |              | poolAddress | string | Yes         | null        | pool address, bech32 string                      |
|         |              | swapRate    | string | Yes         | null        | fis amount = token amount * swapRate, decimals 6 |
|         | swapMaxLimit | N/A         | string | Yes         | null        | the max fis amount limit,  decimals 6            |
|         | swapMinLimit | N/A         | string | Yes         | null        | the min fis amount limit, decimals 6             |


## 2. get swap info

### (1) description

*  get swap info by symbol and txhash

### (2) path

* /feeStation/api/v1/station/swapInfo

### (3) request method

* get

### (4) request param 

* `symbol`: now support `ATOM`
* `txHash`: hex string

### (5) response
* include status、data、message fields
* status、message must be string format,data must be object

| grade 1 | grade 2    | grade 3 | type   | must exist? | encode type | description |
| :------ | :--------- | :------ | :----- | :---------- | :---------- | :---------- |
| status  | N/A        | N/A     | string | Yes         | null        | status code |
| message | N/A        | N/A     | string | Yes         | null        | status info |
| data    | N/A        | N/A     | object | Yes         | null        | data        |
|         | swapStatus | N/A     | number | Yes         | null        | swap status |



* swap status detail

| swap status | description            |
| :---------- | :--------------------- |
| 0           | Default                |
| 1           | TxAlreadySynced        |
| 2           | PayOk                  |
| 3           | AmountLessThanMinLimit |
| 4           | MemoFormatErr          |

