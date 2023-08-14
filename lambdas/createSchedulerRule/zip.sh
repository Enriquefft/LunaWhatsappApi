GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap -tags lambda.norpc main.go
zip CreateScheduleRule.zip bootstrap
# TODO: Upload to aws lambda
