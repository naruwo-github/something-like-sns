> **このドキュメントの目的**: 開発者がローカル環境でプロジェクトを素早く起動するためのクイックスタートガイドです。

## マルチテナントSNS 雛形 README

このリポジトリは、共有DB（行分離）方式のマルチテナントSNSの最小実装の雛形です。
詳細な設計・仕様は設計書（`SOFTWARE_DESIGN.md`）を参照してください。

### 機能概要
- マルチテナント（サブドメイン or `X-Tenant`ヘッダ）
- 認証: Auth0 (OIDC) + 開発用スタブ互換（APIは `X-User` を暫定許可）
- タイムライン、コメント、リアクション（いいね）、DM機能

### 使用スタック
- **サーバ**: Go 1.22+, connect-go, echo, `database/sql`
- **フロント**: Next.js 14+, React 18, React Hooks
- **モノレポ**: Turborepo + pnpm
- **開発環境**: Docker Compose, Makefile

### リポジトリ構成
```
repo/
├─ apps/             # api, web
├─ packages/         # protos, dbschema, shared-ts
├─ infra/            # docker-compose.yml
└─ ...
```

### ローカル起動手順
1. **依存インストール**
   ```bash
   pnpm i
   brew install buf buildifier || true
   ```
2. **生成・DB準備**
   ```bash
   make proto
   docker compose -f infra/local/docker-compose.yml up -d
   make migrate
   make seed
   ```
3. **API/WEB 起動**（別ターミナル）
   ```bash
   make api-dev   # :8080
   make web-dev   # :3000
   ```

### 環境変数（`.env`の例）
```env
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=app
DB_PASS=pass
DB_NAME=sns
API_PORT=8080
ALLOW_DEV_HEADERS=true
NEXT_PUBLIC_API_BASE=http://localhost:8080

# Auth0 (Web: Next.js)
# 参考: https://github.com/auth0/nextjs-auth0
AUTH0_SECRET=replace-with-random-32-bytes-hex
AUTH0_ISSUER_BASE_URL=https://your-tenant.auth0.com
AUTH0_BASE_URL=http://localhost:3000
AUTH0_CLIENT_ID=your-client-id
AUTH0_CLIENT_SECRET=your-client-secret
# APIを保護する場合に取得するトークン用（必要に応じて）
# AUTH0_AUDIENCE=your-api-audience
# AUTH0_SCOPE=openid profile email offline_access
```

### 今後の展望
- **ORMの導入**: `Bun ORM`などの導入によるRepository層の生産性向上。
- **フロントエンドの状態管理**: `TanStack Query`などの導入によるキャッシュ、無限スクロール、楽観的更新の実現。
- **認証**: Web は Auth0 を導入済み。今後は API も JWT/アクセストークン検証を導入し、`X-User` 開発スタブから移行。
