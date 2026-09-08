CREATE TABLE IF NOT EXISTS address (
    id          uuid PRIMARY KEY,
    user_id     varchar(36) NOT NULL,
    street      varchar(255) NOT NULL,
    city        varchar(100) NOT NULL,
    province    varchar(100) NOT NULL,
    postal_code varchar(10) NOT NULL,
    country     varchar(100) NOT NULL,
    is_default  boolean NOT NULL DEFAULT false,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz,
    deleted_at  timestamptz
);

CREATE INDEX ix_address_deleted_at ON address(deleted_at);
CREATE INDEX ix_address_user_id ON address(user_id) WHERE deleted_at IS NULL;
