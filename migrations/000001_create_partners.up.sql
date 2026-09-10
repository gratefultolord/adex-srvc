CREATE TABLE partners (
    uuid UUID PRIMARY KEY,
    name TEXT NOT NULL,
    endpoint TEXT NOT NULL,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,

    countries TEXT[] NOT NULL DEFAULT '{}'::TEXT[],
    device_types TEXT[] NOT NULL DEFAULT '{}'::TEXT[],

    min_bid_floor DOUBLE PRECISION NOT NULL DEFAULT 0,

    blocked_categories TEXT[] NOT NULL DEFAULT '{}'::TEXT[],

    CONSTRAINT partners_name_not_empty
        CHECK (BTRIM(name) <> ''),

    CONSTRAINT partners_endpoint_not_empty
        CHECK (BTRIM(endpoint) <> ''),

    CONSTRAINT partners_min_bid_floor_non_negative
        CHECK (min_bid_floor >= 0)
);