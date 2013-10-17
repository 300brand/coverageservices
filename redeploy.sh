#!/bin/bash
export SKYNET_DZHOST=192.168.20.20:8046

# sky stop -service=WebAPI
# sky stop -service=Social
# sky stop -service=Search
# sky deploy github.com/300brand/coverageservices/WebAPI -host=192.168.20.20
# sky deploy github.com/300brand/coverageservices/Social -host=192.168.20.20
# sky deploy github.com/300brand/coverageservices/Search -host=192.168.20.20
# exit

# go install -v github.com/300brand/coverageservices/{Stats,Article,Feed,Search,Social,StorageReader,StorageWriter,Manager,Queue,WebAPI} || exit

sky stop
sleep 1
# Deploy to all:
for S in Stats Article Feed Search StorageReader StorageWriter Social; do
	sky deploy github.com/300brand/coverageservices/$S
done
# Deploy to sable:
for S in Manager Queue WebAPI; do
#for S in Manager Queue WebAPI Article Feed Search StorageReader StorageWriter; do
	sky deploy github.com/300brand/coverageservices/$S -host=192.168.20.20
done
exit

#sky stop -service=Search
#sky stop -service=StorageReader
#sky stop -service=StorageWriter
#sky stop -service=WebAPI
#sky deploy github.com/300brand/coverageservices/StorageReader -host=192.168.20.20
#sky deploy github.com/300brand/coverageservices/StorageWriter -host=192.168.20.20
#sky deploy github.com/300brand/coverageservices/Search -host=192.168.20.20
#sky deploy github.com/300brand/coverageservices/WebAPI -host=192.168.20.20
