CREATE TABLE IF NOT EXISTS "clips" (
  id            uuid                      PRIMARY KEY DEFAULT uuid_generate_v4(),
  title         varchar                   NOT NULL DEFAULT '',
  "description" varchar                   NOT NULL,
  creator_id    uuid                      REFERENCES "user" (id) NOT NULL,
  created_at    timestamp with time zone  NOT NULL DEFAULT now()
);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_clip_trigram ON "clips" USING gin (f_concat_ws(' ', title, "description") gin_trgm_ops);
