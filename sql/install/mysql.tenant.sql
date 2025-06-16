CREATE TABLE component_acl (
    component   VARCHAR(64)  PRIMARY KEY,                   -- 'content', 'shop'
    enabled     BOOL         NOT NULL DEFAULT TRUE,
    updated_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
                           ON UPDATE CURRENT_TIMESTAMP
);
CREATE INDEX idx_component_acl_enabled ON component_acl (enabled);

CREATE TABLE route_alias (
    alias_path   VARCHAR(255) PRIMARY KEY,
    target_path  VARCHAR(255) NOT NULL,
    created_at   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
                           ON UPDATE CURRENT_TIMESTAMP
);
CREATE INDEX idx_route_alias_updated ON route_alias (updated_at);

CREATE TABLE route_redirect (
    old_path     VARCHAR(255) PRIMARY KEY,
    new_path     VARCHAR(255) NOT NULL,
    created_at   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE role (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    name        VARCHAR(64)  NOT NULL UNIQUE,               -- 'viewer', 'editor'
    enabled     BOOL         NOT NULL DEFAULT TRUE,
    description TEXT,
    created_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
                           ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE role_acl (
    role_id    BIGINT       NOT NULL,
    component  VARCHAR(32)  NOT NULL,
    action     VARCHAR(32)  NOT NULL,                       -- 'view', 'edit'
    permitted  BOOL         NOT NULL DEFAULT TRUE,
    PRIMARY KEY (role_id, component, action),
    FOREIGN KEY (role_id) REFERENCES role(id) ON DELETE CASCADE
);
CREATE INDEX idx_role_acl_lookup ON role_acl (role_id, component);

CREATE TABLE user_role (
    user_id  BIGINT NOT NULL,              -- references global_db.users(id) conceptually
    role_id  BIGINT NOT NULL,              -- FK below
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (role_id) REFERENCES role(id) ON DELETE CASCADE
);

CREATE INDEX idx_user_role_user ON user_role (user_id);