package utils

import "github.com/shopspring/decimal"

// | swap status | description            |
// | :---------- | :--------------------- |
// | 0           | TxNotSynced            |
// | 1           | TxAlreadySynced        |
// | 2           | PayOk                  |
// | 3           | InAmountNotMatch       |
// | 4           | StafiAddressNotMatch   |

const (
	SwapStateNotSynced               = uint8(0)
	SwapStateAlreadySynced           = uint8(1)
	SwapStatePayOk                   = uint8(2)
	SwapStateInAmountNotMatch        = uint8(3)
	SwapStateStafihubAddressNotMatch = uint8(4)
)

var DefaultSwapMaxLimitDeci = decimal.New(100, 6) //default 100e6
var DefaultSwapMinLimitDeci = decimal.New(1, 6)   //default 1e6
var DefaultSwapRateDeci = decimal.New(1, 6)       //default 1e6
