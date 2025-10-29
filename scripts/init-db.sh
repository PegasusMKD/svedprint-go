#!/bin/bash
set -e

# This script runs on PostgreSQL container startup to create multiple databases
# It creates separate databases for each service following the database-per-service pattern

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    -- Create database for the main svedprint service
    SELECT 'CREATE DATABASE svedprint_db'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'svedprint_db')\gexec

    -- Create database for the admin service
    SELECT 'CREATE DATABASE admin_db'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'admin_db')\gexec

    -- Create database for the gateway service (request logging)
    SELECT 'CREATE DATABASE gateway_db'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'gateway_db')\gexec

    -- Create database for Keycloak
    SELECT 'CREATE DATABASE keycloak_db'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'keycloak_db')\gexec

    -- Grant all privileges to the postgres user on all databases
    GRANT ALL PRIVILEGES ON DATABASE svedprint_db TO $POSTGRES_USER;
    GRANT ALL PRIVILEGES ON DATABASE admin_db TO $POSTGRES_USER;
    GRANT ALL PRIVILEGES ON DATABASE gateway_db TO $POSTGRES_USER;
    GRANT ALL PRIVILEGES ON DATABASE keycloak_db TO $POSTGRES_USER;
EOSQL

echo "Databases created successfully:"
echo "  - svedprint_db (main service)"
echo "  - admin_db (admin service)"
echo "  - gateway_db (gateway service)"
echo "  - keycloak_db (authentication)"
