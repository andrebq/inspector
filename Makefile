.PHONY: nix-shell nix-shell-with-host \
	run-dummy-server \
	dist \

nix-shell:
	nix-shell --pure

nix-shell-cmd:
	nix-shell --pure --command "$(command)"

nix-shell-with-host:
	nix-shell

dummyServerPort?=8081
run-dummy-server:
	python -m http.server $(dummyServerPort)

run-load-generator:
	make -C . nix-shell-cmd command="elvish -c 'while true { try { curl -fsSLq http://localhost:8080 2>/dev/null 1>&2 } catch _ignore { nop } ; sleep 1; }'"

dist:
	go build -o ./dist/inspector ./cmd/inspector

docker-build:
	docker build -t andrebq/inspector:latest .

# set push=--push to push after building with buildx
push?=
docker-buildx:
	docker buildx build $(push) --platform linux/amd64 --platform linux/arm64 -t ghcr.io/andrebq/inspector:latest .
