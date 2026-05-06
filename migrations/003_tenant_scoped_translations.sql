-- +goose Up
DROP INDEX IF EXISTS idx_translations_source_hash;
ALTER TABLE translations DROP CONSTRAINT translations_source_hash_source_lang_target_lang_key;
ALTER TABLE translations ADD CONSTRAINT translations_tenant_lang_unique
    UNIQUE (source_hash, source_lang, target_lang, tenant_id);
CREATE INDEX idx_translations_source_hash_tenant ON translations(source_hash, tenant_id);

-- +goose Down
ALTER TABLE translations DROP CONSTRAINT translations_tenant_lang_unique;
ALTER TABLE translations ADD CONSTRAINT translations_source_hash_source_lang_target_lang_key
    UNIQUE (source_hash, source_lang, target_lang);
DROP INDEX IF EXISTS idx_translations_source_hash_tenant;
CREATE INDEX idx_translations_source_hash ON translations(source_hash);
