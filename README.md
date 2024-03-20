# filebeat-otl

The output for filebeat support push events to OTLP，You need to recompile filebeat with the OTLP Output.

# Compile and Build

## We need clone pensieve first
```$xslt
git clone git@github.com:vunetsystems/pensieve.git
```
## the, go to beats directory
```
cd beat/src/github.com/elastic/beats
```
## then, Install OTLP Output, under GOPATH directory
```
go get -u github.com/vikash-vunet/filebeat-otl
```
## modify beats outputs includes, add OTLP output
```
cd {your beats directory}/github.com/elastic/beats/libbeat/publisher/includes/includes.go
```
```
import (
	...
	_ "github.com/vikash-vunet/filebeat-otl"
)
```
## build package, in filebeat
```
cd {your beats directory}/github.com/elastic/beats/filebeat
make
```
 Configure Output
## clickHouse output configuration
```
#----------------------------- ClickHouse output --------------------------------
output.otlp:
  service_name: "vunet"
  service_version: "1.0.0"
  target_url: "http://localhost:8081/uptime"
  oltp_endpoint: "localhost:4317"
  retry_interval: 60
  timeout 300
  bulk_max_size 1000
  max_retries 3
  # will sleep the retry_interval seconds when unexpected exception, default 60s
  retry_interval: 60
```
