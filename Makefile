
.PHONY: redis

redis:
	docker run --name redis -d -p 6379:6379 redis:6.0.9-alpine

.PHONY: run

run:
	go run $(CURDIR)/main.go server

.PHONY: image

image:
	docker build -t vicxu/dcard-work -f $(CURDIR)/build/docker/server.dockerfile .
	docker push vicxu/dcard-work

image.tag:
	docker build -t vicxu/dcard-work:$(tag) -f $(CURDIR)/build/docker/server.dockerfile .
	docker push vicxu/dcard-work:$(tag)

service:
	go build $(CURDIR)/main.go -o service

test:
	@if [ -e $(tag)]