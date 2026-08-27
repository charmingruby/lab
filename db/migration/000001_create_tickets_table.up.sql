CREATE TABLE tickets (
    id VARCHAR(26) PRIMARY KEY,
    title VARCHAR NOT NULL,
    description VARCHAR NOT NULL,
    status VARCHAR NOT NULL,
    priority VARCHAR NOT NULL,
    assignee_id VARCHAR(26) DEFAULT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT NULL,
    deleted_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX tickets_status_idx
    ON tickets (status)
    WHERE deleted_at IS NULL;
