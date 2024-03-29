all:build

get:
	@echo "  >  \033[32mDownloading & Installing all the modules...\033[0m "
	go mod tidy&& go mod download
build:
	@echo "  >  \033[32mBuilding binary...\033[0m "
	cd cmd/stationd && env GOARCH=amd64 go build -o ../../build/stationd
	cd cmd/payerd && env GOARCH=amd64 go build -o ../../build/payerd
	cd cmd/keytool && env GOARCH=amd64 go build -o ../../build/keytool

## license: Adds license header to missing files.
license:
	@echo "  >  \033[32mAdding license headers...\033[0m "
	go get -u github.com/google/addlicense
	addlicense -c "stafiprotocol" -f ./header.txt -y 2021 .

swagger:
	@echo "  >  \033[32mBuilding swagger docs...\033[0m "
	cd cmd/stationd && swag init --parseDependency

fmt:
	go fmt ./...	

clean:
	rm -rf build/

.PHONY: help run build install license