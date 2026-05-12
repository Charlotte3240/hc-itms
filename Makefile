.PHONY: dev dev-backend dev-frontend frontend build clean

dev-backend:
	go run main.go

dev-frontend:
	cd web && npm run dev

frontend:
	cd web && npm install && npm run build

build: frontend
	go build -o hc-itms.exe .

clean:
	rm -rf web/dist hc-itms.exe hc-itms data.db
	rm -f storage/uploads/ios/* storage/uploads/android/* storage/icons/*
