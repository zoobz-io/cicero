-- +goose Up
CREATE TABLE translations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_hash TEXT NOT NULL REFERENCES sources(hash),
    source_lang TEXT NOT NULL,
    target_lang TEXT NOT NULL,
    text TEXT NOT NULL,
    provider TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'completed',
    tenant_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source_hash, source_lang, target_lang)
);

CREATE INDEX idx_translations_source_hash ON translations(source_hash);
CREATE INDEX idx_translations_tenant_id ON translations(tenant_id);

-- +goose Down
DROP TABLE translations;
