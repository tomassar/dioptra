.PHONY: build dev clean ui

build: ui
	go build -o dioptra ./cmd/dioptra

ui:
	cd ui && npm install && npx vite build

dev:
	@echo "Start Go server and Vite dev server separately:"
	@echo "  Terminal 1: go run ./cmd/dioptra connect ..."
	@echo "  Terminal 2: cd ui && npx vite"

clean:
	rm -f dioptra
	rm -rf ui/dist ui/node_modules
