-- Initial schema per 設計書.md

-- tenants
CREATE TABLE IF NOT EXISTS tenants (
  id           BIGINT PRIMARY KEY AUTO_INCREMENT,
  slug         VARCHAR(64) NOT NULL UNIQUE,
  name         VARCHAR(128) NOT NULL,
  created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- tenant_domains
CREATE TABLE IF NOT EXISTS tenant_domains (
  id           BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id    BIGINT NOT NULL,
  domain       VARCHAR(255) NOT NULL UNIQUE,
  created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_tenant_domains_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);

-- users
CREATE TABLE IF NOT EXISTS users (
  id           BIGINT PRIMARY KEY AUTO_INCREMENT,
  auth_sub     VARCHAR(255) NOT NULL UNIQUE,
  display_name VARCHAR(64) NOT NULL,
  avatar_url   VARCHAR(512),
  created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- tenant_memberships
CREATE TABLE IF NOT EXISTS tenant_memberships (
  id           BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id    BIGINT NOT NULL,
  user_id      BIGINT NOT NULL,
  role         ENUM('owner','admin','member') NOT NULL DEFAULT 'member',
  created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uniq_membership (tenant_id, user_id),
  CONSTRAINT fk_memberships_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id),
  CONSTRAINT fk_memberships_user FOREIGN KEY (user_id) REFERENCES users(id)
);

-- posts
CREATE TABLE IF NOT EXISTS posts (
  id             BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id      BIGINT NOT NULL,
  author_user_id BIGINT NOT NULL,
  body           TEXT NOT NULL,
  created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at     TIMESTAMP NULL,
  deleted_at     TIMESTAMP NULL,
  INDEX idx_posts_tenant_created (tenant_id, created_at DESC),
  CONSTRAINT fk_posts_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id),
  CONSTRAINT fk_posts_author FOREIGN KEY (author_user_id) REFERENCES users(id)
);

-- comments
CREATE TABLE IF NOT EXISTS comments (
  id             BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id      BIGINT NOT NULL,
  post_id        BIGINT NOT NULL,
  author_user_id BIGINT NOT NULL,
  body           TEXT NOT NULL,
  created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at     TIMESTAMP NULL,
  INDEX idx_comments_tenant_post_created (tenant_id, post_id, created_at),
  CONSTRAINT fk_comments_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id),
  CONSTRAINT fk_comments_post FOREIGN KEY (post_id) REFERENCES posts(id),
  CONSTRAINT fk_comments_author FOREIGN KEY (author_user_id) REFERENCES users(id)
);

-- reactions
CREATE TABLE IF NOT EXISTS reactions (
  id           BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id    BIGINT NOT NULL,
  target_type  ENUM('post','comment') NOT NULL,
  target_id    BIGINT NOT NULL,
  user_id      BIGINT NOT NULL,
  type         ENUM('like') NOT NULL DEFAULT 'like',
  created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uniq_reaction (tenant_id, target_type, target_id, user_id, type),
  INDEX idx_reactions_tenant_target (tenant_id, target_type, target_id),
  CONSTRAINT fk_reactions_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id),
  CONSTRAINT fk_reactions_user FOREIGN KEY (user_id) REFERENCES users(id)
);

-- conversations
CREATE TABLE IF NOT EXISTS conversations (
  id         BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id  BIGINT NOT NULL,
  kind       ENUM('dm') NOT NULL DEFAULT 'dm',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_conversations_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);

CREATE TABLE IF NOT EXISTS conversation_members (
  id               BIGINT PRIMARY KEY AUTO_INCREMENT,
  conversation_id  BIGINT NOT NULL,
  user_id          BIGINT NOT NULL,
  joined_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uniq_member (conversation_id, user_id),
  CONSTRAINT fk_conv_members_conversation FOREIGN KEY (conversation_id) REFERENCES conversations(id),
  CONSTRAINT fk_conv_members_user FOREIGN KEY (user_id) REFERENCES users(id)
);

-- messages
CREATE TABLE IF NOT EXISTS messages (
  id               BIGINT PRIMARY KEY AUTO_INCREMENT,
  tenant_id        BIGINT NOT NULL,
  conversation_id  BIGINT NOT NULL,
  sender_user_id   BIGINT NOT NULL,
  body             TEXT NOT NULL,
  created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_messages_cnv_created (conversation_id, created_at),
  CONSTRAINT fk_messages_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id),
  CONSTRAINT fk_messages_conversation FOREIGN KEY (conversation_id) REFERENCES conversations(id),
  CONSTRAINT fk_messages_sender FOREIGN KEY (sender_user_id) REFERENCES users(id)
);
