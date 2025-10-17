# Day 2: Database Setup & Initial Migrations

**Phase:** 1 - Project Setup & Database Foundation
**Day:** 2 of 50
**Focus:** PostgreSQL setup, sqlc configuration, create enum and core table migrations

---

## Goals for Today

- [ ] Set up PostgreSQL databases (development and test)
- [ ] Configure sqlc for code generation
- [ ] Create enum type migrations
- [ ] Create school and academic_year table migrations
- [ ] Verify migrations run successfully

---

## Tasks

### Morning Session (4 hours)

#### 1. PostgreSQL Database Setup (1 hour)
- [ ] Verify PostgreSQL 14+ is installed and running
  ```bash
  psql --version
  sudo systemctl status postgresql  # Linux
  ```
- [ ] Create database user for development
  ```sql
  CREATE USER svedprint_dev WITH PASSWORD 'dev_password';
  ALTER USER svedprint_dev CREATEDB;
  ```
- [ ] Create development database
  ```sql
  CREATE DATABASE svedprint_db OWNER svedprint_dev;
  ```
- [ ] Create test database
  ```sql
  CREATE DATABASE svedprint_test_db OWNER svedprint_dev;
  ```
- [ ] Grant necessary privileges
  ```sql
  GRANT ALL PRIVILEGES ON DATABASE svedprint_db TO svedprint_dev;
  GRANT ALL PRIVILEGES ON DATABASE svedprint_test_db TO svedprint_dev;
  ```
- [ ] Test connection
  ```bash
  psql -U svedprint_dev -d svedprint_db -h localhost
  ```
- [ ] Update .env file with actual connection strings

#### 2. sqlc Configuration (30 min)
- [ ] Create sqlc.yaml in project root:
  ```yaml
  version: "2"
  sql:
    - engine: "postgresql"
      queries: "db/queries"
      schema: "db/migrations"
      gen:
        go:
          package: "sqlc"
          out: "db/sqlc"
          sql_package: "pgx/v5"
          emit_json_tags: true
          emit_interface: true
          emit_empty_slices: true
  ```
- [ ] Test sqlc configuration
  ```bash
  sqlc compile
  ```
- [ ] Should show "# package sqlc" or similar

#### 3. Create Enum Migrations (1.5 hours)
- [ ] Create enum migration file
  ```bash
  migrate create -ext sql -dir db/migrations -seq create_enums
  ```
- [ ] Edit `000001_create_enums.up.sql`:
  ```sql
  -- Date type enum for certificate issue dates
  CREATE TYPE date_type AS ENUM ('matriculation', 'certificate');

  -- Gender enum
  CREATE TYPE gender_type AS ENUM ('male', 'female', 'other');

  -- Student conduct/behavior enum
  CREATE TYPE conduct_type AS ENUM ('exemplary', 'good', 'satisfactory', 'unsatisfactory');

  -- Exam period enum
  CREATE TYPE exam_period_type AS ENUM ('june', 'august', 'september');

  -- Student status enum
  CREATE TYPE student_status_type AS ENUM ('regular', 'irregular', 'external');

  -- Matriculation exam type enum
  CREATE TYPE matriculation_type AS ENUM ('state', 'internal');

  -- User type enum (Keycloak roles map to this)
  CREATE TYPE user_type AS ENUM ('teacher', 'admin');
  ```
- [ ] Create down migration `000001_create_enums.down.sql`:
  ```sql
  DROP TYPE IF EXISTS user_type;
  DROP TYPE IF EXISTS matriculation_type;
  DROP TYPE IF EXISTS student_status_type;
  DROP TYPE IF EXISTS exam_period_type;
  DROP TYPE IF EXISTS conduct_type;
  DROP TYPE IF EXISTS gender_type;
  DROP TYPE IF EXISTS date_type;
  ```
- [ ] Test migration
  ```bash
  migrate -path db/migrations -database "postgresql://svedprint_dev:dev_password@localhost:5432/svedprint_db?sslmode=disable" up
  ```
- [ ] Verify enums created in database
  ```sql
  \dT+
  ```
- [ ] Test down migration
  ```bash
  migrate -path db/migrations -database "..." down
  ```
- [ ] Re-run up migration

### Afternoon Session (4 hours)

#### 4. Create School Table Migration (1 hour)
- [ ] Create migration file
  ```bash
  migrate create -ext sql -dir db/migrations -seq create_school
  ```
- [ ] Edit `000002_create_school.up.sql`:
  ```sql
  CREATE TABLE school (
      uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      name VARCHAR(150) NOT NULL,
      administrative_act VARCHAR(40),
      act_year VARCHAR(13),
      director_name VARCHAR(60),
      business_protocol_number VARCHAR(60),
      main_book VARCHAR(5),
      ministry VARCHAR(200),
      country VARCHAR(150) DEFAULT 'Republic of North Macedonia',
      city VARCHAR(150) DEFAULT 'Skopje',
      year_generation VARCHAR(10),
      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );

  -- Indexes
  CREATE INDEX idx_school_name ON school(name);
  CREATE INDEX idx_school_created_at ON school(created_at);
  ```
- [ ] Create down migration `000002_create_school.down.sql`:
  ```sql
  DROP TABLE IF EXISTS school CASCADE;
  ```
- [ ] Run migration
  ```bash
  migrate -path db/migrations -database "..." up
  ```
- [ ] Verify table created
  ```sql
  \d school
  ```

#### 5. Create Academic Year Table Migration (1 hour)
- [ ] Create migration file
  ```bash
  migrate create -ext sql -dir db/migrations -seq create_academic_year
  ```
- [ ] Edit `000003_create_academic_year.up.sql`:
  ```sql
  CREATE TABLE academic_year (
      uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      school_uuid UUID NOT NULL REFERENCES school(uuid) ON DELETE CASCADE,
      name VARCHAR(10) NOT NULL,  -- e.g., "2023/2024"
      certificate_approval_date VARCHAR(75),
      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      CONSTRAINT unique_school_year UNIQUE (school_uuid, name)
  );

  -- Indexes
  CREATE INDEX idx_academic_year_school ON academic_year(school_uuid);
  CREATE INDEX idx_academic_year_name ON academic_year(name);
  CREATE INDEX idx_academic_year_created_at ON academic_year(created_at);
  ```
- [ ] Create down migration `000003_create_academic_year.down.sql`:
  ```sql
  DROP TABLE IF EXISTS academic_year CASCADE;
  ```
- [ ] Run migration and verify

#### 6. Create Certificate Issue Date Table Migration (1 hour)
- [ ] Create migration file
  ```bash
  migrate create -ext sql -dir db/migrations -seq create_certificate_issue_date
  ```
- [ ] Edit `000004_create_certificate_issue_date.up.sql`:
  ```sql
  CREATE TABLE certificate_issue_date (
      uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      school_uuid UUID NOT NULL REFERENCES school(uuid) ON DELETE CASCADE,
      date_type date_type NOT NULL,
      issue_date VARCHAR(40) NOT NULL,
      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );

  -- Indexes
  CREATE INDEX idx_certificate_date_school ON certificate_issue_date(school_uuid);
  CREATE INDEX idx_certificate_date_type ON certificate_issue_date(school_uuid, date_type);
  ```
- [ ] Create down migration
- [ ] Run migration and verify

#### 7. Test All Migrations (30 min)
- [ ] Run all migrations up
  ```bash
  migrate -path db/migrations -database "..." up
  ```
- [ ] Verify all tables exist
  ```sql
  \dt
  ```
- [ ] Check all foreign keys
  ```sql
  SELECT
    tc.table_name,
    kcu.column_name,
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name
  FROM information_schema.table_constraints AS tc
  JOIN information_schema.key_column_usage AS kcu
    ON tc.constraint_name = kcu.constraint_name
  JOIN information_schema.constraint_column_usage AS ccu
    ON ccu.constraint_name = tc.constraint_name
  WHERE tc.constraint_type = 'FOREIGN KEY';
  ```
- [ ] Test down migrations (rollback all)
  ```bash
  migrate -path db/migrations -database "..." down
  ```
- [ ] Verify all tables dropped
  ```bash
  \dt
  ```
- [ ] Re-run up migrations

#### 8. Document Migration Commands (30 min)
- [ ] Add migration section to README.md:
  ```markdown
  ## Database Migrations

  ### Running Migrations
  ```bash
  # Up (apply all pending migrations)
  migrate -path db/migrations -database "${DATABASE_URL}" up

  # Down (rollback one migration)
  migrate -path db/migrations -database "${DATABASE_URL}" down 1

  # Force version (if migrations are out of sync)
  migrate -path db/migrations -database "${DATABASE_URL}" force VERSION
  ```

  ### Creating New Migration
  ```bash
  migrate create -ext sql -dir db/migrations -seq migration_name
  ```
  ```
- [ ] Update .env.example with example DATABASE_URL

---

## Testing

### Verification Steps
- [ ] All migrations run successfully (up)
- [ ] All migrations rollback successfully (down)
- [ ] No orphaned tables after down migrations
- [ ] Foreign key constraints work
- [ ] Enum types created correctly
- [ ] Indexes exist on expected columns

### Manual Tests
- [ ] Insert test record into school table
- [ ] Insert test record into academic_year (with FK to school)
- [ ] Verify CASCADE delete works (delete school, academic_year deleted too)
- [ ] Clean up test data

---

## Documentation

- [ ] Document migration process in README
- [ ] Add database schema diagram (optional, can do later)
- [ ] Note any PostgreSQL version-specific features used
- [ ] Document enum types and their values

---

## Blockers & Issues

**Potential Issues:**
- PostgreSQL not installed or wrong version
- Permission issues creating databases
- Migration tool can't connect to database
- SSL/TLS connection issues
- Enum types not supported (need PostgreSQL 8.3+)

**Solutions:**
- Install PostgreSQL 14+ from official source
- Run psql as postgres superuser initially
- Disable SSL in connection string: `?sslmode=disable`
- Check pg_hba.conf for authentication settings

---

## Tomorrow's Preview

**Day 3 Focus:**
- Create subject_stream and subject table migrations
- Create school_class and class_stream_assignment migrations
- Start student table migration
- Create remaining student-related table migrations

**Preparation:**
- Review DJANGO_TO_GO_MAPPING.md for table schemas
- Understand subject display_order concept
- Review Django array denormalization strategy

---

## Notes

- Using VARCHAR for dates instead of DATE type (matches Django format)
- gen_random_uuid() is built-in to PostgreSQL 13+
- CASCADE deletes ensure referential integrity
- Unique constraints prevent duplicate school/year combinations
- All tables have created_at/updated_at for auditing

---

## Time Tracking

**Estimated:** 8 hours
**Actual:** ___ hours
**Difference:** ___ hours

**Completed?** [ ] Yes [ ] No
**Blockers:** ___________________________
**Notes:** ___________________________
