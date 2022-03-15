module fee-station

go 1.15

require (
	github.com/BurntSushi/toml v0.4.1
	github.com/ChainSafe/go-schnorrkel v0.0.0-20210713215043-76165a18546d
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/btcsuite/btcd v0.22.0-beta
	github.com/cosmos/cosmos-sdk v0.45.1
	github.com/cosmos/ibc-go v1.2.5
	github.com/ethereum/go-ethereum v1.10.15
	github.com/fastly/go-utils v0.0.0-20180712184237-d95a45783239 // indirect
	github.com/gin-gonic/gin v1.6.3
	github.com/go-openapi/spec v0.20.0 // indirect
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869 // indirect
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible
	github.com/lestrrat-go/strftime v1.0.3 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/shopspring/decimal v1.3.1
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/stafihub/cosmos-relay-sdk v0.0.0-20220314073103-34c0e5f10a8f
	github.com/stafihub/stafi-hub-relay-sdk v0.0.0-20220311093735-c6b06c9e7fe9
	github.com/swaggo/gin-swagger v1.3.0
	github.com/swaggo/swag v1.7.0
	github.com/tebeka/strftime v0.1.5 // indirect
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba
	gorm.io/driver/mysql v1.0.3
	gorm.io/gorm v1.21.12
)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

replace github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4
