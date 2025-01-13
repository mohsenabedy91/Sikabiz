-- Table: addresses
CREATE TABLE IF NOT EXISTS addresses
(
    id         INTEGER GENERATED BY DEFAULT AS IDENTITY
        CONSTRAINT pk_addresses PRIMARY KEY,
    street     VARCHAR(255),
    city       VARCHAR(255),
    state      VARCHAR(255),
    zip_code   VARCHAR(255),
    country   VARCHAR(255),
    user_id INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id),
    created_by INTEGER
        CONSTRAINT fk_users_created_by REFERENCES users,
    updated_by INTEGER
        CONSTRAINT fk_users_updated_by REFERENCES users,
    deleted_by INTEGER
        CONSTRAINT fk_users_deleted_by REFERENCES users,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    deleted_at TIMESTAMP WITH TIME ZONE
);
