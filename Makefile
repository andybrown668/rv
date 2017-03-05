VAN_HOST=abrown@192.168.1.51
build: van
	-ssh $(VAN_HOST) pkill van
	rsync -av . $(VAN_HOST):~/
	ssh $(VAN_HOST) sudo ./van

van: *.go
	CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=6 go build van.go

.PHONY: build