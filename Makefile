SHELL := /bin/bash

ifeq ($(shell uname -s),Darwin)
    CGO_CFLAGS := "-I/usr/local/include"
    CGO_LDFLAGS := "-L/usr/local/lib"
endif

ifeq ($(shell uname), Linux)
export LD_LIBRARY_PATH := $(LD_LIBRARY_PATH):/usr/local/lib
endif

.PHONY: furiosa-smi-go-boilerplate
furiosa-smi-go-boilerplate:
	c-for-go -debug -nostamp -out pkg/smi pkg/smi/furiosa-smi.yml
	go build -o clean_cgo tools/clean_cgo.go
	./clean_cgo -source=pkg/smi/binding/types.go -output=pkg/smi/binding/zz_types.go
	./clean_cgo -source=pkg/smi/binding/binding.go -output=pkg/smi/binding/zz_binding.go
	rm -rf pkg/smi/binding/types.go
	rm -rf pkg/smi/binding/binding.go
	rm clean_cgo
