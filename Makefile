# Docker container name for MongoDB
MONGO_CONTAINER="mongodb_pro"

# Database URI (modify if using different credentials)
MONGO_URI="mongodb://mongodb_pro:Dost0n1k@mongo:27017/soand"

# Path to your JavaScript setup script
MONGO_SCRIPT="pkg/scripts/setup_ttl_index.js"

.PHONY: wait-mongo setup-db run

# Wait for MongoDB to be ready
wait-mongo:
	@echo "Waiting for MongoDB to be ready..."
	@until docker exec $(MONGO_CONTAINER) mongosh --eval "db.runCommand({ ping: 1 })" &> /dev/null; do \
		echo "MongoDB is still starting..."; \
		sleep 2; \
	done
	@echo "MongoDB is ready!"

# Task to create the TTL index
setup-db: wait-mongo
	docker exec -i mongodb_pro mongosh "mongodb://mongodb_pro:Dost0n1k@localhost:27017/soand"


# Task to run the application
run:
	go run ./cmd/main.go

.PHONY: run-minio stop-minio restart-minio

# Start MinIO container
run-minio:
	docker run -d --name minio \
	  -p 9000:9000 -p 9001:9001 \
	  -e "MINIO_ROOT_USER=admin" \
	  -e "MINIO_ROOT_PASSWORD=secretpass" \
	  quay.io/minio/minio server /data --console-address ":9001"

# Stop and remove MinIO container
stop-minio:
	docker stop minio || true
	docker rm minio || true

# Restart MinIO container
restart-minio: stop-minio run-minio

swag-gen:
	swag init -g internal/http/comment.go -o docs --parseDependency --parseInternal
	swag init -g internal/http/post.go -o docs --parseDependency --parseInternal
	swag init -g internal/http/user.go -o docs --parseDependency --parseInternal

