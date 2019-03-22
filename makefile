build-wasm:
	GOOS=js GOARCH=wasm go build -o main.wasm main.go
run:
	make build-wasm && go run server.go