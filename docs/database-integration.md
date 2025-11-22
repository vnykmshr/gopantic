# Database Integration Guide

This guide covers how to integrate gopantic with databases, with a focus on PostgreSQL JSONB columns, flexible metadata fields, and common ORM patterns.

## Table of Contents

- [PostgreSQL JSONB Integration](#postgresql-jsonb-integration)
- [Generic Metadata Patterns](#generic-metadata-patterns)
- [ORM Integration](#orm-integration)
- [Validation Strategies](#validation-strategies)
- [Common Patterns](#common-patterns)

## PostgreSQL JSONB Integration

PostgreSQL's JSONB type is perfect for storing flexible, schema-less data. gopantic's `json.RawMessage` support makes JSONB integration seamless.

### Basic JSONB Pattern

```go
import (
    "database/sql"
    "encoding/json"
    "github.com/vnykmshr/gopantic/pkg/model"
)

type Account struct {
    ID          string          `json:"id" validate:"required"`
    Name        string          `json:"name" validate:"required,min=2"`
    Email       string          `json:"email" validate:"required,email"`
    MetadataRaw json.RawMessage `json:"metadata,omitempty"`
    CreatedAt   time.Time       `json:"created_at"`
}

// Database schema
/*
CREATE TABLE accounts (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
*/
```

### Inserting with JSONB

```go
func (r *AccountRepository) Create(input []byte) (*Account, error) {
    // Parse and validate input
    account, err := model.ParseInto[Account](input)
    if err != nil {
        return nil, fmt.Errorf("invalid input: %w", err)
    }

    // Generate ID and timestamp
    account.ID = generateID()
    account.CreatedAt = time.Now()

    // Insert into database
    query := `
        INSERT INTO accounts (id, name, email, metadata, created_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `

    err = r.db.QueryRow(
        query,
        account.ID,
        account.Name,
        account.Email,
        account.MetadataRaw, // PostgreSQL handles json.RawMessage directly
        account.CreatedAt,
    ).Scan(&account.ID)

    if err != nil {
        return nil, fmt.Errorf("database error: %w", err)
    }

    return &account, nil
}
```

### Querying JSONB Data

```go
func (r *AccountRepository) FindByID(id string) (*Account, error) {
    query := `
        SELECT id, name, email, metadata, created_at
        FROM accounts
        WHERE id = $1
    `

    var account Account
    err := r.db.QueryRow(query, id).Scan(
        &account.ID,
        &account.Name,
        &account.Email,
        &account.MetadataRaw,
        &account.CreatedAt,
    )

    if err != nil {
        return nil, err
    }

    return &account, nil
}
```

### JSONB Queries with WHERE Conditions

```go
// Query by JSONB field value
func (r *AccountRepository) FindByTag(tag string) ([]Account, error) {
    query := `
        SELECT id, name, email, metadata, created_at
        FROM accounts
        WHERE metadata @> '{"tags": [{"value": "` + tag + `"}]}'::jsonb
    `

    rows, err := r.db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var accounts []Account
    for rows.Next() {
        var account Account
        err := rows.Scan(
            &account.ID,
            &account.Name,
            &account.Email,
            &account.MetadataRaw,
            &account.CreatedAt,
        )
        if err != nil {
            return nil, err
        }
        accounts = append(accounts, account)
    }

    return accounts, rows.Err()
}
```

### JSONB Updates

```go
func (r *AccountRepository) UpdateMetadata(id string, metadataRaw json.RawMessage) error {
    query := `
        UPDATE accounts
        SET metadata = $1
        WHERE id = $2
    `

    _, err := r.db.Exec(query, metadataRaw, id)
    return err
}

// Partial JSONB update (merge)
func (r *AccountRepository) MergeMetadata(id string, updates json.RawMessage) error {
    query := `
        UPDATE accounts
        SET metadata = COALESCE(metadata, '{}'::jsonb) || $1::jsonb
        WHERE id = $2
    `

    _, err := r.db.Exec(query, updates, id)
    return err
}
```

## Generic Metadata Patterns

### Type-Safe Metadata Access

```go
type Account struct {
    ID          string          `json:"id" validate:"required"`
    Name        string          `json:"name" validate:"required,min=2"`
    MetadataRaw json.RawMessage `json:"metadata,omitempty"`
}

// Define metadata structure
type AccountMetadata struct {
    Preferences Preferences       `json:"preferences"`
    Tags        []string          `json:"tags,omitempty"`
    CustomFields map[string]string `json:"custom_fields,omitempty"`
}

type Preferences struct {
    Theme    string `json:"theme"`
    Language string `json:"language"`
}

// Helper method for type-safe access
func (a *Account) GetMetadata() (*AccountMetadata, error) {
    if len(a.MetadataRaw) == 0 {
        return &AccountMetadata{}, nil
    }

    var metadata AccountMetadata
    if err := json.Unmarshal(a.MetadataRaw, &metadata); err != nil {
        return nil, fmt.Errorf("invalid metadata format: %w", err)
    }

    return &metadata, nil
}

// Usage
account, _ := repo.FindByID("acc_123")
metadata, err := account.GetMetadata()
if err != nil {
    return err
}

fmt.Println(metadata.Preferences.Theme) // Type-safe access
```

### Dynamic Metadata with Validation

```go
type Product struct {
    ID          string          `json:"id" validate:"required"`
    Name        string          `json:"name" validate:"required"`
    SKU         string          `json:"sku" validate:"required,length=8"`
    MetadataRaw json.RawMessage `json:"metadata,omitempty"`
}

func (p *Product) ValidateMetadata() error {
    if len(p.MetadataRaw) == 0 {
        return nil
    }

    var metadata map[string]interface{}
    if err := json.Unmarshal(p.MetadataRaw, &metadata); err != nil {
        return fmt.Errorf("invalid metadata JSON: %w", err)
    }

    // Business rule: if category exists, it must be valid
    if category, ok := metadata["category"].(string); ok {
        validCategories := []string{"electronics", "clothing", "food", "other"}
        if !contains(validCategories, category) {
            return fmt.Errorf("invalid category: %s", category)
        }
    }

    // Business rule: if weight exists, it must be positive
    if weight, ok := metadata["weight"].(float64); ok {
        if weight <= 0 {
            return fmt.Errorf("weight must be positive, got: %f", weight)
        }
    }

    return nil
}

// Two-phase validation pattern
func createProduct(input []byte) (*Product, error) {
    // Phase 1: Structure validation with gopantic
    product, err := model.ParseInto[Product](input)
    if err != nil {
        return nil, err
    }

    // Phase 2: Business logic validation
    if err := product.ValidateMetadata(); err != nil {
        return nil, err
    }

    return &product, nil
}
```

## ORM Integration

### GORM Integration

```go
import (
    "gorm.io/gorm"
    "gorm.io/driver/postgres"
)

type Account struct {
    ID          string          `gorm:"primaryKey" json:"id" validate:"required"`
    Name        string          `gorm:"not null" json:"name" validate:"required,min=2"`
    Email       string          `gorm:"not null;uniqueIndex" json:"email" validate:"required,email"`
    MetadataRaw json.RawMessage `gorm:"type:jsonb" json:"metadata,omitempty"`
    CreatedAt   time.Time       `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt   time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

// Service layer
type AccountService struct {
    db *gorm.DB
}

func (s *AccountService) Create(input []byte) (*Account, error) {
    // Parse and validate with gopantic
    account, err := model.ParseInto[Account](input)
    if err != nil {
        return nil, fmt.Errorf("validation error: %w", err)
    }

    // GORM handles the rest
    if err := s.db.Create(&account).Error; err != nil {
        return nil, fmt.Errorf("database error: %w", err)
    }

    return &account, nil
}

func (s *AccountService) FindByMetadata(key, value string) ([]Account, error) {
    var accounts []Account

    err := s.db.Where(
        "metadata @> ?",
        fmt.Sprintf(`{"%s": "%s"}`, key, value),
    ).Find(&accounts).Error

    return accounts, err
}
```

### sqlx Integration

```go
import (
    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
)

type AccountRepository struct {
    db *sqlx.DB
}

func (r *AccountRepository) Create(input []byte) (*Account, error) {
    // Parse and validate
    account, err := model.ParseInto[Account](input)
    if err != nil {
        return nil, err
    }

    // Insert using sqlx
    query := `
        INSERT INTO accounts (id, name, email, metadata, created_at)
        VALUES (:id, :name, :email, :metadata, :created_at)
        RETURNING id
    `

    account.ID = generateID()
    account.CreatedAt = time.Now()

    rows, err := r.db.NamedQuery(query, &account)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    if rows.Next() {
        rows.Scan(&account.ID)
    }

    return &account, nil
}
```

## Validation Strategies

### Strategy 1: All-in-One (Recommended for Simple Cases)

```go
func CreateAccount(input []byte) (*Account, error) {
    // gopantic does parsing + validation in one step
    account, err := model.ParseInto[Account](input)
    if err != nil {
        return nil, err
    }

    // Save to database
    return repo.Save(account)
}
```

### Strategy 2: Two-Phase (Recommended for Complex Metadata)

```go
func CreateProduct(input []byte) (*Product, error) {
    // Phase 1: Structural validation (gopantic)
    product, err := model.ParseInto[Product](input)
    if err != nil {
        return nil, fmt.Errorf("invalid structure: %w", err)
    }

    // Phase 2: Business logic validation
    if err := product.ValidateMetadata(); err != nil {
        return nil, fmt.Errorf("invalid metadata: %w", err)
    }

    // Phase 3: Database constraints
    return repo.Save(product)
}
```

### Strategy 3: Hybrid (Standard Library + gopantic Validate)

```go
func UpdateAccount(input []byte) (*Account, error) {
    // Use standard json.Unmarshal (handles json.RawMessage perfectly)
    var account Account
    if err := json.Unmarshal(input, &account); err != nil {
        return nil, fmt.Errorf("invalid JSON: %w", err)
    }

    // Apply gopantic validation only
    if err := model.Validate(&account); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    return repo.Update(&account)
}
```

## Common Patterns

### Pattern 1: User Preferences

```go
type User struct {
    ID              string          `json:"id" validate:"required"`
    Email           string          `json:"email" validate:"required,email"`
    PreferencesRaw  json.RawMessage `json:"preferences,omitempty"`
}

type UserPreferences struct {
    Theme            string   `json:"theme"`
    Notifications    bool     `json:"notifications"`
    Language         string   `json:"language"`
    NewsletterTopics []string `json:"newsletter_topics,omitempty"`
}

// Default preferences
func NewUser(email string) *User {
    defaultPrefs := UserPreferences{
        Theme:         "light",
        Notifications: true,
        Language:      "en",
    }

    prefsJSON, _ := json.Marshal(defaultPrefs)

    return &User{
        ID:             generateID(),
        Email:          email,
        PreferencesRaw: prefsJSON,
    }
}
```

### Pattern 2: Audit Metadata

```go
type AuditableEntity struct {
    ID          string          `json:"id" validate:"required"`
    Name        string          `json:"name" validate:"required"`
    MetadataRaw json.RawMessage `json:"metadata,omitempty"`
}

type AuditMetadata struct {
    CreatedBy   string    `json:"created_by"`
    UpdatedBy   string    `json:"updated_by"`
    UpdateCount int       `json:"update_count"`
    LastModified time.Time `json:"last_modified"`
    Notes       []string  `json:"notes,omitempty"`
}

func (e *AuditableEntity) RecordUpdate(userID string, note string) error {
    metadata, err := e.GetMetadata()
    if err != nil {
        metadata = &AuditMetadata{CreatedBy: userID}
    }

    metadata.UpdatedBy = userID
    metadata.UpdateCount++
    metadata.LastModified = time.Now()
    if note != "" {
        metadata.Notes = append(metadata.Notes, note)
    }

    metadataJSON, err := json.Marshal(metadata)
    if err != nil {
        return err
    }

    e.MetadataRaw = metadataJSON
    return nil
}

func (e *AuditableEntity) GetMetadata() (*AuditMetadata, error) {
    if len(e.MetadataRaw) == 0 {
        return nil, fmt.Errorf("no metadata")
    }

    var metadata AuditMetadata
    if err := json.Unmarshal(e.MetadataRaw, &metadata); err != nil {
        return nil, err
    }

    return &metadata, nil
}
```

### Pattern 3: Multi-Tenant Configuration

```go
type TenantConfig struct {
    TenantID    string          `json:"tenant_id" validate:"required"`
    Name        string          `json:"name" validate:"required"`
    ConfigRaw   json.RawMessage `json:"config,omitempty"`
}

type TenantSettings struct {
    Features     map[string]bool `json:"features"`
    Limits       ResourceLimits  `json:"limits"`
    Integrations []Integration   `json:"integrations,omitempty"`
}

type ResourceLimits struct {
    MaxUsers    int `json:"max_users"`
    MaxStorage  int `json:"max_storage_gb"`
    MaxProjects int `json:"max_projects"`
}

type Integration struct {
    Type   string                 `json:"type"`
    Config map[string]interface{} `json:"config"`
}

// Database schema
/*
CREATE TABLE tenant_configs (
    tenant_id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index on JSONB for fast lookups
CREATE INDEX idx_tenant_features ON tenant_configs USING gin (config);
*/
```

### Pattern 4: Event Payload Storage

```go
type Event struct {
    ID          string          `json:"id" validate:"required"`
    Type        string          `json:"type" validate:"required"`
    PayloadRaw  json.RawMessage `json:"payload,omitempty"`
    Timestamp   time.Time       `json:"timestamp"`
}

// Different event payloads
type UserCreatedPayload struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
}

type OrderPlacedPayload struct {
    OrderID string  `json:"order_id"`
    Total   float64 `json:"total"`
    Items   []Item  `json:"items"`
}

// Process events based on type
func ProcessEvent(event *Event) error {
    switch event.Type {
    case "user.created":
        var payload UserCreatedPayload
        if err := json.Unmarshal(event.PayloadRaw, &payload); err != nil {
            return err
        }
        return handleUserCreated(payload)

    case "order.placed":
        var payload OrderPlacedPayload
        if err := json.Unmarshal(event.PayloadRaw, &payload); err != nil {
            return err
        }
        return handleOrderPlaced(payload)

    default:
        return fmt.Errorf("unknown event type: %s", event.Type)
    }
}
```

## Performance Considerations

### JSONB Indexing

```sql
-- GIN index for containment queries (@>)
CREATE INDEX idx_account_metadata ON accounts USING gin (metadata);

-- Specific field indexing
CREATE INDEX idx_account_tags ON accounts USING gin ((metadata->'tags'));

-- Expression index for frequently queried fields
CREATE INDEX idx_account_category ON accounts ((metadata->>'category'));
```

### Query Optimization

```go
// Efficient: Use PostgreSQL's JSONB operators
query := `
    SELECT * FROM accounts
    WHERE metadata @> '{"premium": true}'::jsonb
`

// Less efficient: Fetch all and filter in Go
// Avoid this pattern for large datasets
```

### Caching Metadata Parsing

```go
type CachedAccount struct {
    Account
    metadataCache *AccountMetadata
}

func (ca *CachedAccount) GetMetadata() (*AccountMetadata, error) {
    if ca.metadataCache != nil {
        return ca.metadataCache, nil
    }

    metadata, err := ca.Account.GetMetadata()
    if err != nil {
        return nil, err
    }

    ca.metadataCache = metadata
    return metadata, nil
}
```

## See Also

- [Advanced Type Handling](advanced-types.md) - Complete guide to complex types
- [API Reference](api.md) - Complete API documentation
- [PostgreSQL JSONB Documentation](https://www.postgresql.org/docs/current/datatype-json.html)
- [lib/pq Driver](https://github.com/lib/pq) - PostgreSQL driver for Go
