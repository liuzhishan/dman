#!/bin/sh

cd shared
go build -o dbclient.so -buildmode=c-shared dbclient.go
