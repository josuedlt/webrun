FROM golang:alpine
WORKDIR /src
COPY src .
RUN go mod tidy

### Build current architecture
# RUN go build -ldflags="-s -w" -o build/webrun
# RUN mv build/webrun build/webrun-$(uname -s)-$(uname -m)

### Build by architecture:
RUN GOOS=darwin GOARCH=arm64 go build -o build/webrun-darwin-arm64
RUN GOOS=linux GOARCH=arm64 go build -o build/webrun-linux-arm64
# RUN GOOS=linux GOARCH=arm GOARM=7 go build -o build/webrun-linux-armv7
RUN GOOS=linux GOARCH=amd64 go build -o build/webrun-linux-amd64
# RUN GOOS=windows GOARCH=arm64 go build -o build/webrun-windows-arm64.exe
# RUN GOOS=windows GOARCH=amd64 go build -o build/webrun-windows-amd64.exe

### Enable Compression (takes time)
# RUN apk add upx && upx --brute build/*

RUN chmod +x build/*
ENTRYPOINT [ "/bin/sh", "-c", "cp -R build/* /build" ]