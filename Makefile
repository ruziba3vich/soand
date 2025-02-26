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
