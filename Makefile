docker:
	docker build . -t tibbar/cnskunkworks-operator:latest
docker-push:
	docker push tibbar/cnskunkworks-operator:latest