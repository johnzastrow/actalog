package testhelpers

import (
	"database/sql"
	"fmt"
	"math/rand"
	"regexp"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// SetupTempDB creates a temporary database for Postgres or MySQL and returns
// a DSN that points to the new database, along with a teardown function.
// For sqlite3 the provided DSN is returned unchanged and teardown is a no-op.
func SetupTempDB(driver, dsn string) (string, func(), error) {
	switch driver {
	case "sqlite3":
		return dsn, func() {}, nil
	case "postgres":
		// Replace or add dbname in DSN to connect to postgres admin DB
		re := regexp.MustCompile(`\bdbname=([^ ]+)`)
		adminDSN := dsn
		if re.MatchString(dsn) {
			adminDSN = re.ReplaceAllString(dsn, "dbname=postgres")
		} else {
			adminDSN = dsn + " dbname=postgres"
		}

		adminDB, err := sql.Open("postgres", adminDSN)
		if err != nil {
			return "", nil, err
		}

		dbName := fmt.Sprintf("actalog_test_%d", time.Now().UnixNano()+int64(rand.Intn(1000)))
		if _, err := adminDB.Exec("CREATE DATABASE " + dbName); err != nil {
			adminDB.Close()
			return "", nil, err
		}

		var testDSN string
		if re.MatchString(dsn) {
			testDSN = re.ReplaceAllString(dsn, "dbname="+dbName)
		} else {
			testDSN = dsn + " dbname=" + dbName
		}

		teardown := func() {
			// best-effort cleanup: terminate connections and drop DB
			_ = adminDB.Ping()
			_, _ = adminDB.Exec("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '" + dbName + "' AND pid <> pg_backend_pid();")
			_, _ = adminDB.Exec("DROP DATABASE IF EXISTS " + dbName)
			_ = adminDB.Close()
		}

		return testDSN, teardown, nil

	case "mysql":
		// DSN pattern: user:pass@tcp(host:port)/dbname?params
		re := regexp.MustCompile(`^([^/]+@tcp\([^\)]+\)/)([^?]*)(\?.*)?$`)
		matches := re.FindStringSubmatch(dsn)
		if len(matches) == 0 {
			return "", nil, fmt.Errorf("invalid mysql DSN: %s", dsn)
		}
		prefix := matches[1]
		params := ""
		if len(matches) > 3 {
			params = matches[3]
		}

		adminDSN := prefix + params
		adminDB, err := sql.Open("mysql", adminDSN)
		if err != nil {
			return "", nil, err
		}

		dbName := fmt.Sprintf("actalog_test_%d", time.Now().UnixNano()+int64(rand.Intn(1000)))
		if _, err := adminDB.Exec("CREATE DATABASE `" + dbName + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"); err != nil {
			adminDB.Close()
			return "", nil, err
		}

		testDSN := prefix + dbName + params

		teardown := func() {
			_, _ = adminDB.Exec("DROP DATABASE IF EXISTS `" + dbName + "`")
			_ = adminDB.Close()
		}

		return testDSN, teardown, nil
	default:
		return "", nil, fmt.Errorf("unsupported driver: %s", driver)
	}
}
