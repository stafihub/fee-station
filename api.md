# API doc


## 0. notice

**status code:**

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

**memo:**

```
<uuid hex string>:<stafihubAddress>

example:

5c64203df3974c409b14b49fe7c0bb79:stafi15lne70yk254s0pm2da6g59r82cjymzjqvvqxz7

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
|         |              | decimals    | number | Yes         | null        | native token's decimals                          |
|         |              | poolAddress | string | Yes         | null        | pool address, bech32 string                      |
|         |              | swapRate    | string | Yes         | null        | fis amount = token amount * swapRate, decimals 6 |
|         | swapMaxLimit | N/A         | string | Yes         | null        | the max fis amount limit,  decimals 6            |
|         | swapMinLimit | N/A         | string | Yes         | null        | the min fis amount limit, decimals 6             |
|         | payerAddress | N/A         | string | Yes         | null        | payer address                                    |


## 2. post swap info

### (1) description

*  post user swap info

### (2) path

* /feeStation/api/v1/station/swapInfo

### (3) request method

* post

### (4) request payload 

* data format: application/json
* data detail:

| field           | type   | notice                                                          |
| :-------------- | :----- | :-------------------------------------------------------------- |
| stafihubAddress | string | user stafi address,bech32 string                                |
| symbol          | string | native token like `uatom`, get from pool info                   |
| poolAddress     | string | pool address, get from api                                      |
| inAmount        | string | in token amount, decimal string, decimals equal to native token |
| minOutAmount    | string | min out amount, decimal string, decimals 6                      |

* native token decimals

uatom: 6


### (5) response
* include status、data、message fields
* status、message must be string format, data must be object

| grade 1 | grade 2 | grade 3 | type   | must exist? | encode type | description      |
| :------ | :------ | :------ | :----- | :---------- | :---------- | :--------------- |
| status  | N/A     | N/A     | string | Yes         | null        | status code      |
| message | N/A     | N/A     | string | Yes         | null        | status info      |
| data    | N/A     | N/A     | object | Yes         | null        | data             |
|         | uuid    | N/A     | string | Yes         | null        | uuid, hex string |
       


## 3. get swap info

### (1) description

*  get swap info by uuid

### (2) path

* /feeStation/api/v1/station/swapInfo

### (3) request method

* get

### (4) request param 

* `uuid`: hex string

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

| swap status | description          |
| :---------- | :------------------- |
| 0           | TxNotSynced          |
| 1           | TxAlreadySynced      |
| 2           | PayOk                |
| 3           | InAmountNotMatch     |
| 4           | StafiAddressNotMatch |
