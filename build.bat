@echo off
echo Building dummy logger Go server...
go mod tidy
go build -o dummy-logger-server.exe main.go
echo Build complete! Run with: dummy-logger-server.exe
