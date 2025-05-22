CREATE TABLE `event_outbox` (
  `id`               CHAR(36)    PRIMARY KEY,
  `topic`            VARCHAR(64) NOT NULL,
  `payload`          JSON        NOT NULL,
  `created_at`       TIMESTAMP   NOT NULL DEFAULT NOW(),
  `published_at`     TIMESTAMP   NULL,
  KEY `idx_unpub` (`published_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `site`;
CREATE TABLE `site` (
  `id`            INT UNSIGNED      NOT NULL AUTO_INCREMENT,
  `host`          VARCHAR(256)      NOT NULL,                    -- dev.adept.travel
  `dsn`           VARCHAR(256)      NOT NULL,                    -- tenant DB DSN
  `theme`         VARCHAR(128)      NOT NULL DEFAULT 'base',     -- template set
  `title`         VARCHAR(256)      NOT NULL DEFAULT '',         -- marketing name
  `locale`        VARCHAR(16)       NOT NULL DEFAULT 'en_US',    -- default i18n
  `suspended_at`  TIMESTAMP         DEFAULT NULL,
  `deleted_at`    TIMESTAMP         DEFAULT NULL,
  `created_at`    TIMESTAMP         NOT NULL DEFAULT NOW(),
  `updated_at`    TIMESTAMP         NOT NULL DEFAULT NOW() ON UPDATE NOW(),
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_site_host` (`host`),
  KEY `idx_site_host_active` (`host`, `suspended_at`, `deleted_at`)                -- accelerates ByHost
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;



DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` (
  `id`               INT UNSIGNED AUTO_INCREMENT PRIMARY,
  `username`         VARCHAR(128) NOT NULL,
  `password`         VARCHAR(128) NOT NULL,
  `status`           ENUM('Active', 'Block', 'Inactive', 'Locked') NOT NULL DEFAULT 'Inactive',
  `auth_method`      ENUM('FIDO2', 'Password', 'Token') NOT NULL DEFAULT 'Password',
  `created_at`       TIMESTAMP NOT NULL DEFAULT NOW(),
  `updated_at`       TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),
  `verified_at`      TIMESTAMP DEFAULT NULL,
  `validated_at`     TIMESTAMP DEFAULT NULL,
  `validated_by`     INT DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

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
