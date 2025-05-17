# Authentication Service

## Project Structure

### main.go
Entry point of the application that:
- Implements the gRPC server.
- Contains core authentication logic (CreateUser and LoginUser) 
- Manages service configuration

### models/user.go
Defines the data models:
- User struct with all required fields | Validation tags for data integrity
- Conversion methods between gRPC messages and MongoDB models

### database/mongodb.go
Handles all database connectivity including:
- MongoDB client creation and connection management | Collection access methods
  
### database/init.go
Sets up database schema including:
- Collection schemas and validation rules
- Index creation for query performance
- MongoDB validators for data integrity

### docker-compose.yml
Defines the infrastructure setup:
- MongoDB container configuration
- Network settings
- Environment variables
- Volume mapping for data persistence

## Protocol Buffer Generation

To generate the gRPC service code from the proto definitions, run:

```bash
mkdir -p service/genproto/auth
```
```bash
protoc \
    --proto_path=protobuf "protobuf/auth.proto" \
    --go_out=service/genproto/auth --go_opt=paths=source_relative \
    --go-grpc_out=service/genproto/auth \
    --go-grpc_opt=paths=source_relative
```