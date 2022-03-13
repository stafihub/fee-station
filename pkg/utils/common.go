package utils

import "github.com/shopspring/decimal"

// | swap status | description            |
// | :---------- | :--------------------- |
// | 0           | Default                |
// | 1           | TxAlreadySynced        |
// | 2           | PayOk                  |
// | 3           | AmountLessThanMinLimit |
// | 4           | MemoFormatErr          |

const (
	SwapStateDefault                = uint8(0)
	SwapStateAlreadySynced          = uint8(1)
	SwapStatePayOk                  = uint8(2)
	SwapStateAmountLessThanMinLimit = uint8(3)
	SwapStateMemoFailed             = uint8(4)
)

var DecimalsMap = map[string]int32{
	SymbolAtom: 6,
	SymbolDot:  10,
	SymbolKsm:  12,
	SymbolEth:  18,
}
var DefaultSwapMaxLimitDeci = decimal.New(100, 12) //default 100e12
var DefaultSwapMinLimitDeci = decimal.New(1, 12)   //default 1e12
var DefaultSwapRateDeci = decimal.New(1, 6)        //default 1e6
