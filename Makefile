VAN_IP?=192.168.1.51
VAN_HOST=abrown@$(VAN_IP)

build: van
	-ssh $(VAN_HOST) sudo pkill van
	rsync -av . $(VAN_HOST):~/
	ssh $(VAN_HOST) sudo ./van

van: *.go *.h
	CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=6 go build van.go

test:
	CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=6 go test -c -o van.test van_test.go
	rsync -av . $(VAN_HOST):~/
	ssh $(VAN_HOST) sudo ./van.test -test.bench . -test.cpuprofile cpu.out -test.memprofile mem.out
	scp $(VAN_HOST):~/*.out .
	go tool pprof --pdf van.test cpu.out > cpu.pdf

connect-tty:
	sudo screen /dev/ttyUSB0 115200

.PHONY: build