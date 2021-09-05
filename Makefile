.PHONY: build-and-push
build-and-push:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
	docker build -t taocker/paste:latest .
	docker push taocker/paste:latest
