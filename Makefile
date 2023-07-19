.PHONY: nix-shell nix-shell-with-host \
	run-dummy-server \
	dist \

nix-shell:
	nix-shell --pure

nix-shell-with-host:
	nix-shell

dummyServerPort?=8081
run-dummy-server:
	python -m http.server $(dummyServerPort)

dist:
	go build -o ./dist/inspector .

docker-build:
	docker build -t andrebq/inspector:latest .

push?=
docker-buildx:
	docker buildx build $(push) --platform linux/amd64 --platform linux/arm64 -t ghcr.io/andrebq/inspector:latest .
