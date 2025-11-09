# Multi-Database Support

ActaLog supports three different database systems: SQLite, PostgreSQL, and MySQL/MariaDB. This document explains how to configure and use each database type.

## Supported Databases

- **SQLite 3** - Lightweight, file-based database (recommended for development and small deployments)
- **PostgreSQL 9.6+** - Robust, feature-rich database (recommended for production)
- **MySQL 5.7+ / MariaDB 10.2+** - Popular relational database

## Configuration

Database configuration is managed through environment variables in the `.env` file.

### SQLite Configuration

SQLite is the default database and requires minimal configuration.

```env
DB_DRIVER=sqlite3
DB_NAME=actalog.db
```

**Notes:**
- `DB_NAME` is the file path to the SQLite database file
- No server connection required
- Ideal for development, testing, and single-user deployments
- Supports up to ~10,000 concurrent connections with proper configuration

### PostgreSQL Configuration

For production deployments with PostgreSQL:

```env
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=actalog
DB_PASSWORD=your_secure_password
DB_NAME=actalog
DB_SSLMODE=require
```

**Notes:**
- Requires a running PostgreSQL server
- Recommended for production deployments
- Supports advanced features like full-text search, JSON columns, and concurrent writes
- Better performance for multi-user scenarios

**PostgreSQL Setup:**
```sql
-- Create database
CREATE DATABASE actalog;

-- Create user
CREATE USER actalog WITH PASSWORD 'your_secure_password';

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE actalog TO actalog;
```

### MySQL/MariaDB Configuration

For MySQL or MariaDB deployments:

```env
DB_DRIVER=mysql
DB_HOST=localhost
DB_PORT=3306
DB_USER=actalog
DB_PASSWORD=your_secure_password
DB_NAME=actalog
```

**Notes:**
- Requires a running MySQL or MariaDB server
- Widely supported and compatible
- Good performance for read-heavy workloads
- InnoDB engine used for ACID compliance

**MySQL/MariaDB Setup:**
```sql
-- Create database
CREATE DATABASE actalog CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Create user
CREATE USER 'actalog'@'localhost' IDENTIFIED BY 'your_secure_password';

-- Grant privileges
GRANT ALL PRIVILEGES ON actalog.* TO 'actalog'@'localhost';
FLUSH PRIVILEGES;
```

## Database-Specific Features

### Data Types

ActaLog automatically uses the appropriate data types for each database:

| Concept | SQLite | PostgreSQL | MySQL/MariaDB |
|---------|--------|------------|---------------|
| Auto-increment | INTEGER PRIMARY KEY AUTOINCREMENT | BIGSERIAL PRIMARY KEY | BIGINT AUTO_INCREMENT PRIMARY KEY |
| Text (short) | TEXT | VARCHAR(255) | VARCHAR(255) |
| Text (long) | TEXT | TEXT | TEXT |
| Boolean | INTEGER (0/1) | BOOLEAN | BOOLEAN |
| Float | REAL | DOUBLE PRECISION | DOUBLE |
| Timestamp | DATETIME | TIMESTAMP | DATETIME |

### Schema Management

ActaLog uses a built-in migration system that automatically adapts to the target database:

```go
// Example migration with database-specific SQL
{
    Version:     "0.2.0",
    Description: "Add email_verified column",
    Up: func(db *sql.DB, driver string) error {
        var query string
        switch driver {
        case "sqlite3":
            query = "ALTER TABLE users ADD COLUMN email_verified INTEGER NOT NULL DEFAULT 0"
        case "postgres":
            query = "ALTER TABLE users ADD COLUMN email_verified BOOLEAN NOT NULL DEFAULT FALSE"
        case "mysql":
            query = "ALTER TABLE users ADD COLUMN email_verified BOOLEAN NOT NULL DEFAULT FALSE"
        }
        _, err := db.Exec(query)
        return err
    },
    Down: func(db *sql.DB, driver string) error {
        _, err := db.Exec("ALTER TABLE users DROP COLUMN email_verified")
        return err
    },
}
```

### Indexes

All databases support the same indexes defined in ActaLog:
- Primary key indexes (automatic)
- Foreign key indexes (for relationships)
- Custom indexes on frequently queried columns (email, dates, etc.)

## Migration Between Databases

To migrate from one database to another:

### Export Data

```bash
# From SQLite
sqlite3 actalog.db ".dump" > backup.sql

# From PostgreSQL
pg_dump actalog -U actalog > backup.sql

# From MySQL
mysqldump -u actalog -p actalog > backup.sql
```

### Transform and Import

Due to SQL dialect differences, you may need to transform the dump file before importing. Consider using:
- [pgloader](https://github.com/dimitri/pgloader) - Excellent for SQLite to PostgreSQL
- Custom scripts for data transformation
- ActaLog's data export/import API endpoints (coming soon)

## Performance Considerations

### SQLite

**Pros:**
- Zero configuration
- No separate server process
- Excellent for single-user or low-traffic scenarios
- Very fast for reads

**Cons:**
- Limited concurrent write support
- Single file can grow large
- Not ideal for distributed systems

**Optimization:**
```sql
PRAGMA journal_mode=WAL;
PRAGMA synchronous=NORMAL;
PRAGMA cache_size=-64000;  -- 64MB cache
```

### PostgreSQL

**Pros:**
- Excellent concurrency support
- Advanced indexing (GIN, GiST, etc.)
- Full ACID compliance
- Great for large datasets

**Cons:**
- Requires server setup and maintenance
- More resource-intensive than SQLite

**Optimization:**
- Tune `shared_buffers`, `work_mem`, and `effective_cache_size`
- Use connection pooling (pgBouncer)
- Regular `VACUUM` and `ANALYZE`

### MySQL/MariaDB

**Pros:**
- Widely supported
- Good documentation and tools
- Mature ecosystem

**Cons:**
- Some limitations compared to PostgreSQL
- Different behavior across versions

**Optimization:**
- Use InnoDB engine (default in modern versions)
- Tune `innodb_buffer_pool_size`
- Use query cache appropriately

## Connection Pooling

ActaLog uses Go's `database/sql` package which includes built-in connection pooling:

```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

These values can be tuned based on your deployment needs.

## Testing

ActaLog includes tests that run against all three database types to ensure compatibility. To run tests with a specific database:

```bash
# SQLite (default)
go test ./...

# PostgreSQL
DB_DRIVER=postgres DB_HOST=localhost DB_PORT=5432 DB_USER=test DB_PASSWORD=test DB_NAME=actalog_test go test ./...

# MySQL
DB_DRIVER=mysql DB_HOST=localhost DB_PORT=3306 DB_USER=test DB_PASSWORD=test DB_NAME=actalog_test go test ./...
```

## Troubleshooting

### SQLite

**Error: "database is locked"**
- Enable WAL mode: `PRAGMA journal_mode=WAL;`
- Reduce concurrent writes
- Check for long-running transactions

**Error: "unable to open database file"**
- Check file permissions
- Ensure directory exists and is writable
- Verify DB_NAME path is correct

### PostgreSQL

**Error: "connection refused"**
- Verify PostgreSQL is running
- Check `DB_HOST` and `DB_PORT`
- Ensure `pg_hba.conf` allows connection

**Error: "password authentication failed"**
- Verify DB_USER and DB_PASSWORD
- Check PostgreSQL user permissions
- Ensure database exists

### MySQL/MariaDB

**Error: "Access denied for user"**
- Verify DB_USER and DB_PASSWORD
- Check MySQL user privileges
- Ensure host is correctly specified (`'user'@'localhost'` vs `'user'@'%'`)

**Error: "Unknown database"**
- Verify DB_NAME
- Ensure database was created
- Check case sensitivity

## Best Practices

1. **Development**: Use SQLite for quick iteration and testing
2. **Production**: Use PostgreSQL or MySQL for multi-user deployments
3. **Backups**: Implement regular automated backups regardless of database choice
4. **Security**: Always use strong passwords and enable SSL/TLS for remote connections
5. **Monitoring**: Monitor database performance metrics (connections, query times, etc.)
6. **Migrations**: Test migrations on a copy of production data before applying to production

## Docker Support

ActaLog's docker-compose.yml includes configurations for all three databases. See `docker-compose.yml` for examples.

```bash
# SQLite (no additional services needed)
docker-compose up

# With PostgreSQL
docker-compose --profile postgres up

# With MySQL
docker-compose --profile mysql up
```

## Further Reading

- [SQLite Documentation](https://www.sqlite.org/docs.html)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [MySQL Documentation](https://dev.mysql.com/doc/)
- [Go database/sql Tutorial](https://go.dev/doc/database/overview)
