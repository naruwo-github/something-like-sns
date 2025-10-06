> **このドキュメントの目的**: アプリケーションの技術仕様、アーキテクチャ、データモデル、API定義など、システム全体の詳細設計を網羅的に記述することです。

# マルチテナントSNS 雛形・立ち上げ指示書 v1

この文書 **だけ** を読めば、最小構成のマルチテナントSNS（テナント／ユーザー／投稿／コメント／いいね／DM）を**ローカルで起動**し、**テスト可能**な状態まで立ち上げられます。以降は段階的にクラウド（ECS/Fargate等）へ拡張可能な構造を前提とします。

---

## 0. スコープ / 非スコープ

* **目的**: 共有DB（行分離）方式のマルチテナントSNSを最小実装し、API/DB/FE が一体で動く雛形を提供。
* **含む**: テナント解決、認証スタブ、CRUD/API、DM（2者DM）、無限スクロール、簡易レート制限、シードデータ、E2Eテスト。
* **含まない**: 画像アップロード、課金/請求、通知配信、監査証跡の厳密化、本番運用のセキュリティ強化（WAF/脆弱性診断等）。

---

## 1. 使用スタック（必須/代替）

* **サーバ**（必須）: Go 1.22+ / **connect-go**（RPC）/ **echo**（ヘルス/静的）/ **`database/sql`** (MySQL) / golang-migrate
* **フロント**（必須）: Next.js 14+（App Router）/ React 18 / **React Hooks (`useState`/`useEffect`)**
* **モノレポ**（必須）: Turborepo + pnpm
* **開発環境**（必須）: Docker Compose（MySQL, Adminer）/ Makefile
* **観測**（任意）: OpenTelemetry SDK（ローカルexporter）
* **代替（サーバ）**: TypeScript + Express + Prisma（設計は同一）。 ※Goが難しい場合のみ。

> **理由**: 本番は Go × スキーマ駆動を想定。雛形では TS 代替も用意可能だが、基本は Go 実装とする。

---

## 2. リポジトリ構成（モノレポ）

```
repo/
├─ apps/
│  ├─ api/                 # Go API (connect-go)
│  └─ web/                 # Next.js フロント
├─ packages/
│  ├─ protos/              # .proto と生成設定（buf）
│  ├─ dbschema/            # DDL・マイグレーション（golang-migrate）
│  └─ shared-ts/           # Webで共有する型/ユーティリティ
├─ infra/
│  └─ local/               # docker-compose.yml / seed スクリプト
├─ README.md               # Quick start guide
├─ SOFTWARE_DESIGN.md      # Detailed software design
├─ AWS_ARCHITECTURE.md     # AWS infrastructure design
├─ Makefile
├─ turbo.json
└─ package.json (pnpm)
```

---

## 3. マルチテナント方針

* **方式**: 共有DB（行分離）。**全テーブルに `tenant_id`** を持たせ、**全クエリで `WHERE tenant_id = :ctx_tenant`** を強制。
* **テナント解決**: `Host` ヘッダのサブドメイン、または開発用 `X-Tenant` ヘッダ → `tenant_id` を解決。
* **スコープ強制**: サーバの **Request Middleware** で `ctx.tenant_id` を注入。**DAO/Repository は ctx 必須**で、`tenant_id` 条件が無いクエリを拒否。

---

## 4. 権限モデル

* **ロール**: `owner` / `admin` / `member`
* **可視性**: 同一 `tenant_id` に属するデータのみアクセス可。
* **操作権限**（最小）:

  * 投稿/コメント作成: `member+`
  * 削除: 作成者 or `admin+`
  * いいね: `member+`
  * DM: 同一テナントのユーザー同士のみ

---

## 5. データモデル（DDL）

> エンジン: InnoDB / 文字コード: utf8mb4 / `created_at` は `DEFAULT CURRENT_TIMESTAMP`

```sql
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
```

> **DM 2者制約**（アプリ側）: `GetOrCreateDM(a,b)` 実装で、`a < b` の正規化キー（例: `direct_key = sha1(min(a,b) || ':' || max(a,b))`）を conversation 拡張列として UNIQUE 付与してもよい（雛形ではアプリ側でユニーク性を担保）。

---

## 6. API 設計（Connect RPC / gRPC 互換）

> フロントは Connect-Web クライアントを利用。**全APIはサーバ側で `tenant_id` を注入**し、クライアントから渡させない。

### 6.1 .proto 構成

```
packages/protos/
├─ buf.yaml
├─ buf.gen.yaml            # Go/TSコード生成設定
└─ sns/v1/
   ├─ tenant.proto
   ├─ timeline.proto
   ├─ reaction.proto
   └─ dm.proto
```

### 6.2 サービス定義（抜粋）

```proto
// sns/v1/tenant.proto
syntax = "proto3";
package sns.v1;
option go_package = "github.com/example/repo/gen/sns/v1;v1";

message ResolveTenantRequest { string host = 1; }
message ResolveTenantResponse { uint64 tenant_id = 1; string slug = 2; }
message GetMeRequest {}
message GetMeResponse {
  uint64 user_id = 1;
  string display_name = 2;
  repeated TenantMembership memberships = 3;
}
message TenantMembership { uint64 tenant_id = 1; string role = 2; string tenant_slug = 3; }

service TenantService {
  rpc ResolveTenant(ResolveTenantRequest) returns (ResolveTenantResponse);
  rpc GetMe(GetMeRequest) returns (GetMeResponse);
}
```

```proto
// sns/v1/timeline.proto
syntax = "proto3";
package sns.v1;
option go_package = "github.com/example/repo/gen/sns/v1;v1";

message Cursor { string token = 1; }
message Post {
  uint64 id = 1; uint64 author_user_id = 2; string body = 3; string created_at = 4; bool liked_by_me = 5; uint32 like_count = 6; uint32 comment_count = 7;
}
message Comment { uint64 id = 1; uint64 post_id = 2; uint64 author_user_id = 3; string body = 4; string created_at = 5; }

message ListFeedRequest { Cursor cursor = 1; }
message ListFeedResponse { repeated Post items = 1; Cursor next = 2; }
message CreatePostRequest { string body = 1; }
message CreatePostResponse { Post post = 1; }
message ListCommentsRequest { uint64 post_id = 1; Cursor cursor = 2; }
message ListCommentsResponse { repeated Comment items = 1; Cursor next = 2; }
message CreateCommentRequest { uint64 post_id = 1; string body = 2; }
message CreateCommentResponse { Comment comment = 1; }

service TimelineService {
  rpc ListFeed(ListFeedRequest) returns (ListFeedResponse);
  rpc CreatePost(CreatePostRequest) returns (CreatePostResponse);
  rpc ListComments(ListCommentsRequest) returns (ListCommentsResponse);
  rpc CreateComment(CreateCommentRequest) returns (CreateCommentResponse);
}
```

```proto
// sns/v1/reaction.proto
syntax = "proto3";
package sns.v1;
option go_package = "github.com/example/repo/gen/sns/v1;v1";

enum TargetType { TARGET_TYPE_UNSPECIFIED = 0; POST = 1; COMMENT = 2; }
message ToggleReactionRequest { TargetType target_type = 1; uint64 target_id = 2; string type = 3; }
message ToggleReactionResponse { bool active = 1; uint32 total = 2; }

service ReactionService { rpc ToggleReaction(ToggleReactionRequest) returns (ToggleReactionResponse); }
```

```proto
// sns/v1/dm.proto
syntax = "proto3";
package sns.v1;
option go_package = "github.com/example/repo/gen/sns/v1;v1";

message Conversation { uint64 id = 1; string created_at = 2; repeated uint64 member_user_ids = 3; }
message Message { uint64 id = 1; uint64 conversation_id = 2; uint64 sender_user_id = 3; string body = 4; string created_at = 5; }

message GetOrCreateDMRequest { uint64 other_user_id = 1; }
message GetOrCreateDMResponse { uint64 conversation_id = 1; }
message ListConversationsRequest { Cursor cursor = 1; }
message ListConversationsResponse { repeated Conversation items = 1; Cursor next = 2; }
message ListMessagesRequest { uint64 conversation_id = 1; Cursor cursor = 2; }
message ListMessagesResponse { repeated Message items = 1; Cursor next = 2; }
message SendMessageRequest { uint64 conversation_id = 1; string body = 2; }
message SendMessageResponse { Message message = 1; }

service DMService {
  rpc GetOrCreateDM(GetOrCreateDMRequest) returns (GetOrCreateDMResponse);
  rpc ListConversations(ListConversationsRequest) returns (ListConversationsResponse);
  rpc ListMessages(ListMessagesRequest) returns (ListMessagesResponse);
  rpc SendMessage(SendMessageRequest) returns (SendMessageResponse);
}
```

**カーソル**: `token` には `base64("created_at:id")` 等を入れ、`created_at DESC, id DESC` の複合ソートで安定ページング。

---

## 7. サーバ実装規約（Go）

* **Context 必須**: `ctx` から `tenant_id` / `user_id` を取得。`ctx` を受けない Repo 関数は禁止。
* **ミドルウェア**: `ResolveTenant` → `Authenticate(スタブ可)` → `Inject(ctx)` の順で実行。
* **`database/sql`**: クエリでは `tenant_id` を明示。SQLインジェクション対策を徹底。
* **エラーハンドリング**: gRPC status codes に準拠（`InvalidArgument/Unauthenticated/PermissionDenied/NotFound/AlreadyExists`）。
* **トランザクション**: 整合性が必要な複数クエリは `sql.Tx` を使用。
* **バリデーション**: サーバ側で body 長/空チェック。最大 2000 文字。

---

## 8. フロント実装規約（Next.js）

* **データ取得**: React Hooks (`useState`, `useEffect`) と `fetch` API で、Connect-Webエンドポイントを直接呼び出し。
* **UI**: `/`=Feed、`/dm`=会話一覧、`/dm/[id]`=メッセージ。Tailwindで簡素に。
* **テナント/ユーザー**: 現在は各コンポーネントでハードコード。

---

## 9. 認証スタブ

* 開発では `X-User` ヘッダを許容し、`users.auth_sub` に紐づくユーザーに偽装ログイン。
* 初回は存在しなければ作成し、`tenant_memberships` に `member` で自動参加（シードに依存）。

---

## 10. レート制限（最小）

* **投稿**: 1ユーザー 1分あたり 10 回
* **コメント**: 1ユーザー 1分あたり 20 回
* **DM送信**: 1ユーザー 1分あたり 20 回
* 実装: メモリ内トークンバケット（将来は Redis へ差し替え）。

---

## 11. ローカル起動手順（必須）

1. 依存インストール

```
# ルート
pnpm i
# Go ツール
brew install buf buildifier || true
```

2. 生成（proto → Go/TS）

```
make proto
```

3. DB 起動 & マイグレーション & シード

```
docker compose -f infra/local/docker-compose.yml up -d
make migrate
make seed
```

4. API/WEB 起動

```
# 別ターミナル
make api-dev   # apps/api を起動（:8080）
make web-dev   # apps/web を起動（:3000）
```

5. 動作確認

* ブラウザで `http://acme.localhost:3000/` → フィードが表示され、投稿/コメント/いいねが可能
* `http://acme.localhost:3000/dm` → 会話一覧、`/dm/{id}` で送受信可能

---

## 12. 環境変数（.env.example）

```
# DB
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=app
DB_PASS=pass
DB_NAME=sns

# API
API_PORT=8080
ALLOW_DEV_HEADERS=true       # X-Tenant / X-User を許容

# WEB
NEXT_PUBLIC_API_BASE=http://localhost:8080
```

---

## 13. シードデータ（最低）

* `tenants`: `acme` / `beta`
* `users`: `u_alice`, `u_bob`, `u_caro`
* `tenant_memberships`: 全員 `acme` に参加、`alice=owner`, `bob=admin`, `caro=member`
* `posts`: 5件、`comments`: 各2件、`reactions`: ランダム付与
* `dm`: `alice-bob` の会話＋メッセージ 3件

---

## 14. 受け入れ条件（Acceptance Criteria）

1. **テナント隔離**: `acme.localhost` のデータが `beta.localhost` で見えない（E2Eテストで検証）。
2. **フィード**: 投稿→即時反映。無限スクロールが `created_at DESC, id DESC` で安定。
3. **コメント**: 投稿詳細でコメント一覧/作成が可能。
4. **いいね**: 1ユーザー1対象1種類のみトグル。総数が正しく反映。
5. **DM**: `GetOrCreateDM` で2者DMが1会話に収束。送受信可能。
6. **権限**: 削除は作者 or admin+ のみ。
7. **レート制限**: 閾値超過で 429 を返却。
8. **リント/フォーマット/ビルド**: `pnpm -w lint && pnpm -w build` が通る。

---

## 15. テスト方針

* **ユニット**: Repo/Usecase に対する in-memory or transaction rollback テスト。
* **API**: サーバ立ち上げた上での結合テスト（`ListFeed/CreatePost/ToggleReaction`）。
* **E2E（web）**: Playwright でサブドメイン差し替えテスト（acme ↔ beta）。

---

## 16. 運用・拡張の足場（将来）

* **ORMの導入**: `Bun ORM` などの導入による、Repository層のクエリビルドの安全性・生産性の向上。
* **フロントエンドの状態管理**: `TanStack Query` などの導入による、キャッシュ、無限スクロール、楽観的更新などの実現。
* **認証**: `middleware.ts` を利用したテナント解決や、本格的な認証（例: OIDC）の導入。
* **監査ログ**: 重要操作のみイベントテーブル（`audit_events`）へ追記。
* **メディア**: S3直PUT + 署名URL。`attachments` テーブルを追加。
* **通知**: Webhook・メール・WS は後日。
* **本番**: ECS/Fargate へ移行、RDS(Aurora MySQL)、Secrets Manager、OTel → Datadog exporter。

---

## 17. 責務分担（役割ごとのToDo）

* **Backend**: DDL適用、proto定義→生成、Repository/Usecase/Service 実装、シード、APIテスト。
* **Frontend**: 初回 ResolveTenant、認証スタブ、Feed/Detail/DM 画面、Query キャッシュ、E2E。
* **Infra**: docker-compose、Makefile、CI（lint/build/test）、Playwright 走行環境準備。

---

## 18. 作業チェックリスト

* [ ] `docker compose up -d` で DB/Adminer が起動する
* [ ] `make migrate && make seed` が成功
* [ ] `make proto` で Go/TS 生成物が更新される
* [ ] `make api-dev` で 8080 が Listen しヘルスが 200
* [ ] `make web-dev` で 3000 にアクセスできる
* [ ] フィード/コメント/いいね/DM の基本動作が通る
* [ ] `pnpm -w test` がグリーン

---

## 19. 命名・コード規約（抜粋）

* **Go**: `internal/` に Usecase/Repo を配置。`pkg/` は公開ユーティリティのみ。
* **TS/React**: Server Components を基本、クライアント要素はフォーム/無限スクロールのみ。
* **コミット**: Conventional Commits（`feat:`, `fix:`）。
* **フォーマット**: gofmt / golangci-lint / ESLint / Prettier。

---

## 20. 付録：最小 Make ターゲット例

```Makefile
proto:
	cd packages/protos && buf generate

migrate:
	migrate -path packages/dbschema/migrations -database "mysql://${DB_USER}:${DB_PASS}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}" up

seed:
	cd apps/api && go run ./cmd/seed/main.go

api-dev:
	cd apps/api && (GO111MODULE=on go run github.com/air-verse/air@v1.52.2 || go run ./cmd/server)

web-dev:
	pnpm --filter @repo/web dev
```

## バックエンドアーキテクチャ

バックエンドAPIは、ヘキサゴナルアーキテクチャ（別名：ポーツ＆アダプター）の原則に従って設計されています。この設計の目的は、アプリケーションの中核をなすビジネスロジック（`domain`層、`application`層）を、データベースやRPCフレームワークといった外部の技術的関心事から明確に分離することです。

### ディレクトリ構造

-   `/cmd/server`: アプリケーションのメインエントリーポイント。
-   `/internal/domain`: 中核となるビジネスエンティティ（例: `Post`, `User`）を定義します。外部依存のない純粋なデータ構造です。
-   `/internal/port`: アプリケーションコアとアダプタ層の境界となるインターフェース（Port）を定義します。UsecaseやRepositoryのインターフェースが含まれます。
-   `/internal/application`: Usecaseインターフェースを実装します。ビジネスロジックの調整役であり、この層がアプリケーションの動作を記述します。
-   `/internal/adapter`: Portを実装する具体的なアダプタを配置します。
    -   `/handler/rpc`: RPCリクエストを処理し、Usecaseを呼び出す入力アダプタ。
    -   `/repository/mysql`: RepositoryインターフェースをMySQLで実装する出力アダプタ。

### トランザクション管理 (Unit of Work パターン)

このアーキテクチャにおける重要なルールの一つが、**「データベーストランザクションの境界は、Usecase層（Application層）が管理する」**という点です。

このルールを徹底するため、**Unit of Work パターン**を採用し、`port.Store`インターフェースを介して実装しています。

-   **なぜこのルールが必要か？**
    一つのビジネスオペレーション（Usecase）が、複数のリポジトリへの書き込みを必要とするケースは少なくありません（例：DMを作成し、同時に通知も送信する）。その一連の処理が「すべて成功するか、すべて失敗するかのどちらか」であるべき（原子性を持つべき）だと知っているのは、ビジネスロジック全体を把握しているUsecase層だけです。トランザクションの制御をUsecaseに委ねることで、複数のデータ操作にまたがるデータ整合性を保証できます。

-   **どのように実装しているか？**
    1.  Usecaseは、個々のリポジトリではなく`port.Store`インターフェースに依存します。
    2.  トランザクションが必要な処理では、Usecaseは`store.ExecTx(ctx, func(txStore port.Store) error { ... })`を呼び出します。
    3.  `ExecTx`に渡された関数の中で、引数の`txStore`から取得したリポジトリへの操作は、すべて単一のデータベーストランザクション内で実行されます。
    4.  `Store`の実装が、関数から返される`error`の有無に応じて、自動的に`COMMIT`または`ROLLBACK`を実行します。

このパターンにより、将来的に複雑なビジネスロジックが追加された場合でも、アーキテクチャの基本設計を変更することなく、安全に機能を拡張できる堅牢な基盤を構築しています。