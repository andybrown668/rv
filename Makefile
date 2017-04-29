VAN_IP?=192.168.1.51
VAN_HOST=abrown@$(VAN_IP)

build: stop van
	rsync -av . $(VAN_HOST):~/
	ssh $(VAN_HOST) ./van

van: *.go *.h
	CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=6 go build -o $@ app/main.go

test:
	CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=6 go test -c -o van.test van_test.go
	rsync -av . $(VAN_HOST):~/
	ssh $(VAN_HOST) sudo ./van.test -test.bench . -test.cpuprofile cpu.out -test.memprofile mem.out
	scp $(VAN_HOST):~/*.out .
	go tool pprof --pdf van.test cpu.out > cpu.pdf

grab:
	rm -f *.jpg
	rsync -av $(VAN_HOST):~/*.jpg .

stop:
	ssh $(VAN_HOST) "pkill van || true"

ssh:
	ssh $(VAN_HOST)

connect-tty:
	sudo screen /dev/ttyUSB0 115200

.PHONY: build