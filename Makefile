build_contract_bin:
	solc --optimize --bin ./contracts/contracts.sol -o build

build_contract_abi:
	solc --optimize --abi ./contracts/contracts.sol -o build

build_contract_api_go: 
	abigen --abi=./build/BalanceCheck.abi --bin=./build/BalanceCheck.bin --pkg=api --out=./api/contracts.go
