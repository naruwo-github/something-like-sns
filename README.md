## マルチテナントSNS 雛形 README

このリポジトリは、共有DB（行分離）方式のマルチテナントSNSの最小実装の雛形です。
詳細な設計・仕様は設計書（`設計書.md`）を参照してください。

### 機能概要
- マルチテナント（サブドメイン or `X-Tenant`ヘッダ）
- 開発用の認証スタブ (`X-User`ヘッダ)
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
```

### 今後の展望
- **ORMの導入**: `Bun ORM`などの導入によるRepository層の生産性向上。
- **フロントエンドの状態管理**: `TanStack Query`などの導入によるキャッシュ、無限スクロール、楽観的更新の実現。
- **認証**: `middleware.ts`を利用したテナント解決や、本格的な認証の導入。