#! /bin/bash

protoc --gogofaster_out=plugins=grpc:. *.proto
#protoc --go_out=plugins=grpc:. *.proto
