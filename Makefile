.PHONY: dev dev-backend dev-frontend build clean

dev:
	@make -j2 dev-backend dev-frontend

dev-backend:
	cd backend && air

dev-frontend:
	cd frontend && npm run dev

build:
	cd frontend && npm run build
	cd backend && go build -o ../bin/mcp-server-manager ./cmd/server

clean:
	rm -rf bin/ frontend/.nuxt frontend/.output frontend/node_modules
