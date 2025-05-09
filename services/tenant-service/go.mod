module github.com/rsturla/dmarc-monitor/services/tenant-service

go 1.22.5

require github.com/aws/aws-lambda-go v1.47.0 // indirect

// Rewrite the github.com/rsturla/dmarc-monitor/libs/common module to the local path (../../libs/common)
replace github.com/rsturla/dmarc-monitor/libs/common => ../../libs/common

require (
	github.com/rsturla/dmarc-monitor/libs/common v0.0.0
)
