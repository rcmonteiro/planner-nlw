# History

```bash
# Install dependencies
go install github.com/discord-gophers/goapi-gen@latest

# Generate specs from OpenAPI spec
goapi-gen --out ./internal/api/spec/planner.gen.spec.go ./internal/api/spec/planner.spec.json

# It adds any missing module requirements necessary to build the current module’s packages and dependencies
go mod tidy 

# Update modules to the latest version
go get -u ./...

# Run db with docker
docker-compose up -d

# Install Tern to help with SQL migrations
go install github.com/jackc/tern/v2@latest
tern init ./internal/pgstore/migrations/
tern new --migrations ./internal/pgstore/migrations/ create_trips_table
tern new --migrations ./internal/pgstore/migrations/ create_participants_table
tern new --migrations ./internal/pgstore/migrations/ create_activities_table
tern new --migrations ./internal/pgstore/migrations/ create_links_table

# After writing all the migrations, run them
tern migrate --migrations ./internal/pgstore/migrations/ --config ./internal/pgstore/migrations/tern.conf

# Generate code based on the OpenAPI spec
go generate ./...

# After writing the queries, we can create the code running:
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
sqlc generate -f ./internal/pgstore/sqlc.yaml
go mod tidy 
go get -u ./...

# Add the new command to the gen.go, then run to check if it works
go generate ./...
```
