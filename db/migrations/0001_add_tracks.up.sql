CREATE TABLE track_imports
(
    id           BIGSERIAL PRIMARY KEY,
    owner_id     TEXT                        NOT NULL,
    inserted_at  TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITHOUT TIME ZONE,
    failed_at    TIMESTAMP WITHOUT TIME ZONE,
    error        TEXT,
    filename     TEXT                        NOT NULL,
    data         BYTEA                       NOT NULL
);

CREATE TABLE tracks
(
    id          BIGSERIAL PRIMARY KEY,
    owner_id    TEXT,
    name        TEXT,
    upload_time TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    time        TIMESTAMP WITHOUT TIME ZONE,
    geojson     JSONB                       NOT NULL,
    import_id   BIGINT REFERENCES track_imports (id)
);
