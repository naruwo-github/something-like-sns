## マルチテナントSNS 雛形 README

このリポジトリは、共有DB（行分離）方式のマルチテナントSNS（テナント／ユーザー／投稿／コメント／いいね／DM）をローカルで起動・テストできる最小実装の雛形です。詳細は設計書（`設計書.md`）を参照してください。

### 機能サマリ（仕様）
- **テナント解決**: `Host` のサブドメイン、または開発用 `X-Tenant` ヘッダから `tenant_id` を解決
- **認証スタブ（開発）**: `X-User` ヘッダで `users.auth_sub` に紐づくユーザーとして偽装ログイン（未存在なら作成）
- **権限ロール**: `owner` / `admin` / `member`
- **タイムライン**: 投稿作成、無限スクロール（`created_at DESC, id DESC` 安定ページング）
- **コメント**: 投稿詳細で一覧/作成
- **リアクション**: いいね（Post/Commentに対し1ユーザー1種類をトグル）
- **DM**: 2者DMのみ、`GetOrCreateDM` で1会話に収束
- **レート制限（最小）**: 投稿 10/分、コメント 20/分、DM送信 20/分（メモリ内トークンバケット）
- **マルチテナント隔離**: すべてのデータアクセスは `tenant_id` でスコープ強制

### 推奨スタック
- **サーバ**: Go 1.22+, connect-go（RPC）, echo（ヘルス/静的）, Bun ORM, MySQL 8.0, golang-migrate
- **フロント**: Next.js 14+（App Router）, React 18, TanStack Query, Tailwind CSS
- **モノレポ**: Turborepo + pnpm
- **開発環境**: Docker Compose（MySQL, Adminer）, Makefile
- **観測（任意）**: OpenTelemetry（ローカルexporter）

### リポジトリ構成（想定）
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
├─ turbo.json
├─ package.json (pnpm)
├─ Makefile
└─ README.md
```

### マルチテナント方針
- **共有DB（行分離）**: 全テーブルに `tenant_id` を持ち、全クエリで `WHERE tenant_id = :ctx_tenant` を強制
- **テナント解決**: リクエスト `Host` のサブドメイン、または開発用 `X-Tenant` ヘッダ
- **スコープ注入**: サーバのミドルウェアで `ctx.tenant_id` を注入。Repository/DAO は `ctx` 必須

### API 概要（Connect RPC / gRPC互換）
- フロントは Connect-Web クライアントを使用。`tenant_id` はサーバが注入（クライアントから渡さない）
- 主なサービス
  - `TenantService`: `ResolveTenant`, `GetMe`
  - `TimelineService`: `ListFeed`, `CreatePost`, `ListComments`, `CreateComment`
  - `ReactionService`: `ToggleReaction`
  - `DMService`: `GetOrCreateDM`, `ListConversations`, `ListMessages`, `SendMessage`
- **カーソル**: `Cursor.token = base64("created_at:id")` で安定ページング（`created_at DESC, id DESC`）

### データモデル（要点）
- 主なテーブル: `tenants`, `tenant_domains`, `users`, `tenant_memberships`, `posts`, `comments`, `reactions`, `conversations`, `conversation_members`, `messages`
- 外部鍵/ユニーク制約で整合性を担保。リアクションはポリモーフィック（Post/Comment）
- DMは2者のみ。アプリ側で1会話に収束するよう実装

### ローカル起動手順（抜粋）
1. 依存インストール
   ```bash
   pnpm i
   brew install buf buildifier || true
   ```
2. 生成（proto → Go/TS）
   ```bash
   make proto
   ```
3. DB起動・マイグレーション・シード
   ```bash
   docker compose -f infra/local/docker-compose.yml up -d
   make migrate
   make seed
   ```
4. API/WEB 起動（別ターミナル）
   ```bash
   make api-dev   # :8080
   make web-dev   # :3000
   ```
5. 動作確認
   - `http://acme.localhost:3000/` でフィード表示、投稿/コメント/いいねが可能
   - `http://acme.localhost:3000/dm` → 会話一覧、`/dm/{id}` で送受信可能

### 環境変数（.env の例）
```env
# DB
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=app
DB_PASS=pass
DB_NAME=sns

# API
API_PORT=8080
ALLOW_DEV_HEADERS=true

# WEB
NEXT_PUBLIC_API_BASE=http://localhost:8080
```

### シードデータ（最低）
- `tenants`: `acme`, `beta`
- `users`: `u_alice`, `u_bob`, `u_caro`
- `tenant_memberships`: 全員 `acme` に参加（`alice=owner`, `bob=admin`, `caro=member`）
- `posts`: 5件、`comments`: 各2件、`reactions`: ランダム付与
- `dm`: `alice-bob` の会話＋メッセージ3件

### 受け入れ条件（抜粋）
1. テナント隔離（`acme.localhost` と `beta.localhost` のデータが混在しない）
2. フィードは投稿直後に反映、無限スクロールが安定
3. コメント一覧/作成が可能
4. いいねは1ユーザー1対象1種類のトグルで総数整合
5. DMは2者で1会話に収束し送受信可能
6. 削除は作者または `admin+` のみ
7. レート制限超過で 429 を返却
8. `pnpm -w lint && pnpm -w build` が通る

### テスト方針（要約）
- ユニット: Repo/Usecase の in-memory または Tx ロールバック
- API結合: サーバを立てたうえで `ListFeed/CreatePost/ToggleReaction` 等
- E2E: Playwright でサブドメイン差し替えテスト（acme ↔ beta）

### フロント実装メモ
- `middleware.ts` で Host から `tenantSlug` 抽出し、初回 `ResolveTenant` 実行
- データ取得は TanStack Query（`staleTime: 30s`）、無限スクロールは `useInfiniteQuery`
- いいねは楽観的更新（失敗時リバート）

詳細・DDL・proto定義は `設計書.md` を参照してください。


