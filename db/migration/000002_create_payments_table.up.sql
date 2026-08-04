CREATE TABLE payments (
    id VARCHAR(26) PRIMARY KEY,
    user_id VARCHAR(26) NOT NULL,
    offering_id VARCHAR(26) NOT NULL REFERENCES offerings (id),
    status VARCHAR NOT NULL,
    charged_amount INTEGER NOT NULL,
    external_id VARCHAR NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT NULL,
    deleted_at TIMESTAMP DEFAULT NULL
);