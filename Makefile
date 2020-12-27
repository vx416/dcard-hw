
.PHONY: redis

redis:
	docker run --name redis -d -p 6379:6379 redis:6.0.9-alpine

.PHONY: run

run:
	DATA_PATH=$(CURDIR)/configs/animals.json go run $(CURDIR)/main.go server

.PHONY: image

image:
	docker build -t vicxu/dcard-work -f $(CURDIR)/build/docker/server.dockerfile .
	docker push vicxu/dcard-work

service:
	go build $(CURDIR)/main.go -o service

app.yaml:
	cp $(CURDIR)/configs/app-dev.yaml $(CURDIR)/configs/app.yaml

.PHONY: deploy

deploy: ansible.cfg hosts image
	ansible-playbook $(CURDIR)/deployments/ansible/server-playbook.yaml

.PHONY: ping.production

ping.production:
	curl https://209.97.172.117.nip.io/