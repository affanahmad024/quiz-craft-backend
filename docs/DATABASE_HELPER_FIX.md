# Database Helper Fix

## Add GetSQLDB() Method

You need to add this method to your `pkg/database/database.go` file to expose the underlying `*sql.DB` for the DAO layer.

### Location
File: `c:\Users\affan\OneDrive\Documents\quizMaster\quiz-craft\backend\pkg\database\database.go`

### Add This Function

```go
// GetSQLDB returns the underlying *sql.DB from GORM
// This is needed for the DAO layer which uses raw SQL
func GetSQLDB() *sql.DB {
	if gormDB == nil {
		return nil
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Errorf("Failed to get *sql.DB from GORM: %v", err)
		return nil
	}

	return sqlDB
}
```

### Why This Is Needed

The new DAO layer uses raw SQL queries instead of GORM. It needs direct access to `*sql.DB` to execute queries with:
- `db.QueryRowContext()`
- `db.QueryContext()`
- `db.ExecContext()`

GORM wraps `*sql.DB`, so we need to extract it using `gormDB.DB()`.

### Full Example

If your `database.go` looks like this:

```go
package database

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var gormDB *gorm.DB

func InitPostgres(cfg *config.DatabaseConfig) error {
	dsn := fmt.Sprintf(...)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	gormDB = db
	return nil
}

func Close() {
	if gormDB != nil {
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	}
}
```

Add this method:

```go
// Add this function
func GetSQLDB() *sql.DB {
	if gormDB == nil {
		return nil
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil
	}

	return sqlDB
}
```

### Import Required

Make sure you have this import at the top:

```go
import (
	"database/sql"
	// ... other imports
)
```

### That's It!

After adding this method, your `main.go` will be able to call:

```go
sqlDB := database.GetSQLDB()
userdao.Init(sqlDB)
```

And the DAO layer will have direct SQL access! ✅
