-- 001_create_raw_events.sql (PostgreSQL)

CREATE TABLE IF NOT EXISTS raw_events (
    id              BIGSERIAL PRIMARY KEY,
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    provider_code   VARCHAR(50)  NOT NULL,
    data_code       VARCHAR(50)  NOT NULL,
    waybill_number  VARCHAR(100),
    tracking_number VARCHAR(100),
    customer_code   VARCHAR(100),
    track_node_code VARCHAR(100),
    process_time    VARCHAR(50),
    payload         TEXT         NOT NULL,
    envelope_meta   TEXT,
    status          VARCHAR(20)  NOT NULL DEFAULT 'pending',
    retry_count     INT          NOT NULL DEFAULT 0,
    max_retries     INT          NOT NULL DEFAULT 5,
    last_error      TEXT,
    processed_at    TIMESTAMP    NULL,
    created_at      TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_raw_events_status  ON raw_events(status);
CREATE INDEX IF NOT EXISTS idx_raw_events_waybill ON raw_events(waybill_number);
CREATE INDEX IF NOT EXISTS idx_raw_events_created ON raw_events(created_at);
CREATE INDEX IF NOT EXISTS idx_raw_events_retry   ON raw_events(status, retry_count, updated_at);

-- tracking_details
CREATE TABLE IF NOT EXISTS tracking_details (
    id                BIGSERIAL PRIMARY KEY,
    tracking_number   VARCHAR(100) NOT NULL UNIQUE,
    service_class     VARCHAR(100),
    status            INT          NOT NULL DEFAULT 0,
    detail            JSONB,
    last_detail       JSONB,
    synced_at         TIMESTAMP    NOT NULL DEFAULT NOW(),
    error_message     TEXT,
    tracking_log_id   BIGINT,
    auto_delivered_at TIMESTAMP    NULL,
    created_at        TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tracking_details_log_id ON tracking_details(tracking_log_id);

-- tracking_logs
CREATE TABLE IF NOT EXISTS tracking_logs (
    id                     BIGSERIAL PRIMARY KEY,
    tracking_number        VARCHAR(100),
    source_tracking_number VARCHAR(100),
    channel_alias          VARCHAR(100),
    shipping_agent         VARCHAR(100),
    shipping_channel       VARCHAR(100),
    country_code           VARCHAR(10),
    track_status           INT          NOT NULL DEFAULT 0,
    synced_at              TIMESTAMP    NULL,
    received_at            TIMESTAMP    NULL,
    delivered_at           TIMESTAMP    NULL,
    tracked_at             TIMESTAMP    NULL,
    fulfill_at             TIMESTAMP    NULL,
    created_at             TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMP    NOT NULL DEFAULT NOW()
);
