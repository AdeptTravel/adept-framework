CREATE TABLE event_outbox (
  id            CHAR(36)    PRIMARY KEY,
  topic         VARCHAR(64) NOT NULL,
  payload       JSON        NOT NULL,
  created_at    TIMESTAMP   NOT NULL DEFAULT NOW(),
  published_at  TIMESTAMP  NULL,
  KEY idx_unpub (published_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;