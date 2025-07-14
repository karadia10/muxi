# Muxi - Spring Boot style Go HTTP Framework

**Features**
- Annotation-based handler discovery (`// @route /foo [GET]`)
- Codegen for route registration
- Type-safe, ergonomic handler signatures
- Zero-config, net/http-powered

**Quickstart**

```sh
cd muxi
go mod tidy
cd example
go run ../discover.go > muxi_autoroutes.go
go run main.go
