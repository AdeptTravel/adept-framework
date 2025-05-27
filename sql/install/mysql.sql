CREATE TABLE `event_outbox` (
  `id`               CHAR(36)    PRIMARY KEY,
  `topic`            VARCHAR(64) NOT NULL,
  `payload`          JSON        NOT NULL,
  `created_at`       TIMESTAMP   NOT NULL DEFAULT NOW(),
  `published_at`     TIMESTAMP   NULL,
  KEY `idx_unpub` (`published_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS site;

CREATE TABLE site (
  id            BIGINT PRIMARY KEY AUTO_INCREMENT,
  host          VARCHAR(256)  NOT NULL UNIQUE,
  dsn           VARCHAR(512)  NOT NULL,
  theme         VARCHAR(128)  NOT NULL DEFAULT 'base',
  title         VARCHAR(256)  NOT NULL DEFAULT '',
  locale        VARCHAR(16)   NOT NULL DEFAULT 'en_US',
  routing_mode  VARCHAR(6)    NOT NULL DEFAULT 'path',
  route_version INT           NOT NULL DEFAULT 0,

  suspended_at  TIMESTAMP NULL,
  deleted_at    TIMESTAMP NULL,
  created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Optional check constraint that both engines respect.
ALTER TABLE site
  ADD CONSTRAINT ck_site_routing_mode
  CHECK (routing_mode IN ('path', 'mapped', 'mixed'));




-- ---------- core user record ----------
CREATE TABLE users (
  id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  username        VARCHAR(128) NOT NULL,
  email           VARCHAR(256) NOT NULL,
  status          ENUM('Active', 'Blocked', 'Inactive', 'Locked') NOT NULL DEFAULT 'Inactive',
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  verified_at     TIMESTAMP NULL,
  -- application-level bookkeeping
  UNIQUE KEY uk_users_username (username),
  UNIQUE KEY uk_users_email    (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE password_credentials (
  id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id         BIGINT UNSIGNED NOT NULL,
  password_hash   VARBINARY(255) NOT NULL,                   -- Argon2id or bcrypt
  needs_rotation  TINYINT(1)     NOT NULL DEFAULT 0,
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  UNIQUE KEY uk_pwd_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE oauth_identities (
  id                 BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id            BIGINT UNSIGNED NOT NULL,
  provider           ENUM('Google', 'GitHub', 'Microsoft', 'Apple', 'Custom') NOT NULL,
  provider_user_id   VARCHAR(256) NOT NULL,                  -- sub claim or user ID from IdP
  access_token_enc   VARBINARY(2048) NULL,                   -- optional, encrypted at rest
  refresh_token_enc  VARBINARY(2048) NULL,
  expires_at         DATETIME NULL,
  created_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  UNIQUE KEY uk_oauth_provider_uid (provider, provider_user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE passkey_credentials (
  id                 BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id            BIGINT UNSIGNED NOT NULL,
  credential_id      VARBINARY(255) NOT NULL,                -- base64url decoded bytes
  public_key         VARBINARY(512) NOT NULL,                -- COSE key bytes
  sign_count         BIGINT UNSIGNED NOT NULL DEFAULT 0,
  transports         SET('usb', 'nfc', 'ble', 'internal') NOT NULL,
  created_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_used_at       TIMESTAMP NULL,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  UNIQUE KEY uk_passkey_credid (credential_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `user_profile`;
CREATE TABLE `user_profile` (
  `id`               INT UNSIGNED AUTO_INCREMENT,
  `user_id`          INT UNSIGNED NOT NULL,
  `name_prefix`      VARCHAR(40) DEFAULT NULL,
  `name_first`       VARCHAR(160) NOT NULL,
  `name_middle`      VARCHAR(160) DEFAULT NULL,
  `name_last`        VARCHAR(160) NOT NULL,
  `name_suffix`      VARCHAR(40) DEFAULT NULL,
  `name_display`     VARCHAR(320) DEFAULT NULL,
  `dob`              DATE DEFAULT NULL,
  `status`           ENUM('Active', 'Block', 'Inactive', 'Locked') NOT NULL DEFAULT 'Inactive',
  `created_at`       TIMESTAMP NOT NULL DEFAULT NOW(),
  `updated_at`       TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
