CREATE TABLE offerings (
    id VARCHAR(26) PRIMARY KEY,
    name VARCHAR NOT NULL,
    description VARCHAR NOT NULL,
    charge_type VARCHAR NOT NULL,
    price INTEGER NOT NULL,
    currency VARCHAR NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT NULL,
    deleted_at TIMESTAMP DEFAULT NULL
);

CREATE UNIQUE INDEX offerings_name_active_unique_idx
    ON offerings (name)
    WHERE deleted_at IS NULL;
