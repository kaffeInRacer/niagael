CREATE TABLE IF NOT EXISTS promo_usage (
    id          uuid PRIMARY KEY,
    promo_id    uuid NOT NULL REFERENCES promo(id),
    user_id     uuid NOT NULL,
    quantity    int NOT NULL DEFAULT 0,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz,
    CONSTRAINT ux_promo_usage UNIQUE (promo_id, user_id)
);

CREATE INDEX ix_promo_usage_promo_id ON promo_usage(promo_id);
CREATE INDEX ix_promo_usage_user_id ON promo_usage(user_id);
