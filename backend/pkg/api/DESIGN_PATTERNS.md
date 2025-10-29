# Backend Design Patterns and Architecture

This document describes the design patterns and architectural decisions used in the PolarDBX UI Backend.

## Table of Contents
1. [Architectural Overview](#architectural-overview)
2. [Design Patterns](#design-patterns)
3. [Code Organization](#code-organization)
4. [Usage Examples](#usage-examples)

## Architectural Overview

The backend follows a layered architecture:

```
┌─────────────────────────────────────┐
│   HTTP Layer (Gin Handlers)        │
│   - Routes, Middleware              │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│   API Layer                         │
│   - Handler Factories               │
│   - Request/Response Utilities      │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│   Domain Layer (Optional)           │
│   - Business Logic                  │
│   - Orchestration                   │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│   Repository Layer (K8s Package)    │
│   - Kubernetes Client Operations    │
│   - Resource CRUD                   │
└─────────────────────────────────────┘
```

## Design Patterns

### 1. Factory Pattern

**Location**: `pkg/api/handler/crud_factory.go`

**Purpose**: Eliminate repetitive CRUD handler code by generating handlers from resource operations.

**Benefits**:
- Reduces ~500+ lines of duplicated code
- Type-safe generic handlers using Go generics
- Consistent error handling and response formatting
- Easy to test and maintain

**Example**:
```go
// Define resource operations
type SystemTaskOps struct{}

func (ops SystemTaskOps) List(ctx context.Context, cli client.Client, namespace string) ([]polardbxv1.SystemTask, error) {
    return k8s.ListSystemTasksWithContext(ctx, cli, namespace)
}

// Create factory
factory := handler.NewCRUDHandlerFactory(
    SystemTaskOps{},
    "system task",
    func() *polardbxv1.SystemTask { return &polardbxv1.SystemTask{} },
)

// Use in routes
r.GET("", factory.List())
r.POST("", factory.Create())
r.GET("/:namespace/:name", factory.Get())
r.PUT("/:namespace/:name", factory.Update())
r.DELETE("/:namespace/:name", factory.Delete())
```

### 2. Builder Pattern

**Location**: `pkg/api/util/response.go`

**Purpose**: Provide a fluent interface for building HTTP responses consistently.

**Benefits**:
- Consistent response format across all endpoints
- Flexible response building with method chaining
- Better API documentation through standardization

**Example**:
```go
// Complex response with builder
util.NewResponse().
    Status(http.StatusOK).
    Data(result).
    Message("Operation successful").
    Send(c)

// Simple convenience functions
util.Success(c, data)
util.Created(c, newResource)
util.BadRequest(c, "Invalid input", validationErr.Error())
```

### 3. Middleware Pattern

**Location**: `pkg/api/handlers.go`

**Purpose**: Extract cross-cutting concerns like authentication and client initialization.

**Benefits**:
- Centralized authentication logic
- Request-scoped Kubernetes clients
- Audit trail with user context

**Example**:
```go
// Middleware chain
router := gin.Default()
v1 := router.Group("/api/v1")
v1.Use(api.KubeconfigAuthMiddleware())
```

### 4. Repository Pattern

**Location**: `pkg/k8s/client.go`

**Purpose**: Abstract Kubernetes client operations and provide a clean API for resource management.

**Benefits**:
- Centralized K8s client logic
- Easy to mock for testing
- Consistent error handling

**Example**:
```go
// Repository functions
func ListSystemTasksWithContext(ctx context.Context, cli client.Client, namespace string) ([]polardbxv1.SystemTask, error)
func CreateSystemTaskWithContext(ctx context.Context, cli client.Client, namespace string, obj *polardbxv1.SystemTask) (*polardbxv1.SystemTask, error)
```

### 5. Utility/Helper Pattern

**Location**: `pkg/api/util/http.go`

**Purpose**: Centralize common operations like context extraction and error handling.

**Benefits**:
- DRY (Don't Repeat Yourself)
- Consistent error mapping to HTTP status codes
- Simplified handler code

**Example**:
```go
// Extract client from context
cli, ok := util.K8sClientFromContext(c)
if !ok {
    return // Error response already sent
}

// Handle K8s errors consistently
if err := k8s.DeleteResource(ctx, cli, ns, name); err != nil {
    util.HandleK8sError(c, "failed to delete resource", err)
    return
}
```

## Code Organization

### Package Structure

```
backend/
├── pkg/
│   ├── api/
│   │   ├── handler/          # Handler factories (NEW)
│   │   │   └── crud_factory.go
│   │   ├── crd/              # CRD route aliases
│   │   │   ├── common/       # Common route helpers (NEW)
│   │   │   └── */routes.go   # Resource-specific routes
│   │   ├── domain/           # Domain/business logic
│   │   ├── util/             # Utilities
│   │   │   ├── http.go       # HTTP helpers
│   │   │   ├── response.go   # Response builders (NEW)
│   │   │   └── security.go   # Security utilities
│   │   ├── handlers.go       # Middleware
│   │   └── router/           # Route registration
│   ├── k8s/                  # Kubernetes client wrapper
│   └── config/               # Configuration
```

### Design Principles

1. **Separation of Concerns**: Each layer has a specific responsibility
2. **DRY (Don't Repeat Yourself)**: Use factories and utilities to avoid duplication
3. **Single Responsibility**: Each package/module has one clear purpose
4. **Open/Closed Principle**: Open for extension, closed for modification
5. **Dependency Inversion**: Depend on abstractions (interfaces) not concretions

## Usage Examples

### Example 1: Creating a New CRUD Resource Handler

```go
// 1. Define your resource operations (implements ResourceOperations interface)
type MyResourceOps struct{}

func (ops MyResourceOps) List(ctx context.Context, cli client.Client, namespace string) ([]MyResource, error) {
    // Implementation
}
// ... implement other CRUD methods

// 2. Create the factory
factory := handler.NewCRUDHandlerFactory(
    MyResourceOps{},
    "my resource",
    func() *MyResource { return &MyResource{} },
)

// 3. Register routes
func RegisterRoutes(crd *gin.RouterGroup) {
    common.RegisterNamespaceScopedCRUD(crd, "myresources", common.CRUDHandlers{
        List:   factory.List(),
        Create: factory.Create(),
        Get:    factory.Get(),
        Update: factory.Update(),
        Delete: factory.Delete(),
    })
}
```

### Example 2: Custom Endpoint with Response Builders

```go
func MyCustomEndpoint(c *gin.Context) {
    cli, ok := util.K8sClientFromContext(c)
    if !ok {
        return
    }
    
    // Your business logic
    result, err := performComplexOperation(cli)
    if err != nil {
        util.BadRequest(c, "operation failed", err.Error())
        return
    }
    
    // Return success
    util.Success(c, result)
}
```

### Example 3: Using Route Helpers

```go
func RegisterRoutes(crd *gin.RouterGroup) {
    // Standard CRUD routes
    r := common.RegisterNamespaceScopedCRUD(crd, "resources", common.CRUDHandlers{
        List:   List,
        Create: Create,
        Get:    Get,
        Update: Update,
        Delete: Delete,
    })
    
    // Additional custom routes on the same resource
    item := r.Group("/:namespace/:name")
    item.GET("/status", GetStatus)
    item.POST("/action", PerformAction)
}
```

## Benefits Summary

### Code Quality Improvements
- **~650+ lines of code eliminated** through factories and helpers
- **Consistent error handling** across all endpoints
- **Type-safe** operations using Go generics
- **Better testability** with clear separation of concerns

### Maintainability Improvements
- **Single source of truth** for CRUD operations
- **Easy to extend** with new resources
- **Clear patterns** for new developers to follow
- **Reduced cognitive load** with standardized code

### Performance
- **No runtime overhead** - patterns are compile-time constructs
- **Efficient memory usage** with proper context handling
- **Connection pooling** through singleton K8s clients

## Migration Guide

For existing handlers that haven't been migrated to the new patterns:

1. **Identify the resource type** and its CRUD operations
2. **Create a resource operations struct** implementing `ResourceOperations[T]`
3. **Use the factory** to generate handlers
4. **Update routes** to use the common route helpers
5. **Test thoroughly** to ensure behavior is preserved

## Future Enhancements

Potential improvements for consideration:

1. **Pagination support** in List operations
2. **Filtering and sorting** standardization
3. **Request validation** middleware
4. **Rate limiting** patterns
5. **Caching strategies** for read-heavy operations
6. **OpenAPI/Swagger** generation from handlers
7. **Observability** (metrics, tracing) middleware
