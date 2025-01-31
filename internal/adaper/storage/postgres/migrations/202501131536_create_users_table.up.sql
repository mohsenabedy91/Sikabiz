-- Table: users
CREATE TABLE IF NOT EXISTS users
(
    id           INTEGER GENERATED BY DEFAULT AS IDENTITY
        CONSTRAINT pk_users PRIMARY KEY,
    uuid         uuid                     DEFAULT gen_random_uuid() UNIQUE,
    first_name   VARCHAR(128),
    last_name    VARCHAR(128),
    email        VARCHAR(128) UNIQUE,
    phone_number VARCHAR(128) UNIQUE,
    created_by   INTEGER
        CONSTRAINT fk_users_created_by REFERENCES users,
    updated_by   INTEGER
        CONSTRAINT fk_users_updated_by REFERENCES users,
    deleted_by   INTEGER
        CONSTRAINT fk_users_deleted_by REFERENCES users,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at   TIMESTAMP WITH TIME ZONE DEFAULT now(),
    deleted_at   TIMESTAMP WITH TIME ZONE
);

-- Index: idx_users_email
CREATE INDEX IF NOT EXISTS idx_users_email
    ON users (email);

-- Index: idx_users_phone_number
CREATE INDEX IF NOT EXISTS idx_users_phone_number
    ON users (phone_number);