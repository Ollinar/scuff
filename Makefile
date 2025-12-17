DB_FILE=test/test.db
SQL_SCRIPT=schema.sql

# Default target
all: syncdb

# Run SQL script into the database
syncdb:
	sqlite3 $(DB_FILE) < $(SQL_SCRIPT)

# Clean target (optional) - remove the database file
clean:
	rm -f $(DB_FILE)

build:
	go build -o build.bin ./cmd/web.go 

