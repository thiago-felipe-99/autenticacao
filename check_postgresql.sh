#!/bin/sh

# PostgreSQL server connection details
PG_HOST="localhost"
PG_PORT="5432"
PG_USER="postgres"
PG_PASSWORD="postgres"
PG_DATABASE="postgres"

# Function to check if PostgreSQL has initialized
check_postgresql_initialized() {
    psql -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d "$PG_DATABASE" -c "SELECT 1;" > /dev/null 2>&1
    return $?
}

# Wait for PostgreSQL to initialize
wait_for_postgresql() {
    MAX_TRIES=30
    SLEEP_INTERVAL=1
    TRIES=0

    while [ $TRIES -lt $MAX_TRIES ]; do
        check_postgresql_initialized
        if [ $? -eq 0 ]; then
            echo "PostgreSQL has initialized successfully."
            return 0
        fi

        echo "PostgreSQL is still initializing. Retrying in $SLEEP_INTERVAL seconds..."
        sleep $SLEEP_INTERVAL
        TRIES=$((TRIES + 1))
    done

    echo "PostgreSQL initialization failed after $MAX_TRIES attempts."
    return 1
}

wait_for_postgresql

