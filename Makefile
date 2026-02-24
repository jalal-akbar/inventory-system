# Variabel
BINARY_NAME=rabella-app
MAIN_PATH=./cmd/main.go
PORT=7070

## run: Build and run the application directly for cleaner process management
run: build
	@echo "Starting application..."
	./$(BINARY_NAME)

## build: Compile the binary
build:
	@echo "Building application..."
	go build -o $(BINARY_NAME) $(MAIN_PATH)

## build: Kompilasi menjadi file .exe untuk Windows (buat Kakak)
build-win:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME).exe $(MAIN_PATH)

## seed: Menjalankan database seeder
seed:
	@echo "Running database seeder..."
	go run cmd/seed/main.go

## clean: Membersihkan file sementara dan database (hati-hati!)
clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME).exe
	-@fuser -k $(PORT)/tcp || true

## reset: Menghapus database dan membuatnya ulang dari nol
reset-db:
	@echo "Shutting down any lingering processes..."
	-@fuser -k $(PORT)/tcp || true
	@echo "Removing database file..."
	rm -f inventory.db inventory.db-shm inventory.db-wal
	@echo "Database has been nuked! Ready for fresh run."