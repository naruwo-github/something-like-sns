package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func getenv(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}

func must(err error, msg string) {
    if err != nil {
        log.Fatalf("%s: %v", msg, err)
    }
}

func openDB() *sql.DB {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true&charset=utf8mb4,utf8",
        getenv("DB_USER", "app"), getenv("DB_PASS", "pass"), getenv("DB_HOST", "127.0.0.1"), getenv("DB_PORT", "3306"), getenv("DB_NAME", "sns"))
    db, err := sql.Open("mysql", dsn)
    must(err, "open db")
    must(db.Ping(), "ping db")
    return db
}

func upsertTenant(db *sql.DB, slug, name string) (int64, error) {
    _, err := db.Exec("INSERT INTO tenants (slug, name) VALUES (?, ?) ON DUPLICATE KEY UPDATE name=VALUES(name)", slug, name)
    if err != nil { return 0, err }
    var id int64
    err = db.QueryRow("SELECT id FROM tenants WHERE slug=?", slug).Scan(&id)
    return id, err
}

func upsertTenantDomain(db *sql.DB, tenantID int64, domain string) error {
    _, err := db.Exec("INSERT INTO tenant_domains (tenant_id, domain) VALUES (?, ?) ON DUPLICATE KEY UPDATE tenant_id=VALUES(tenant_id)", tenantID, domain)
    return err
}

func upsertUser(db *sql.DB, authSub, displayName string) (int64, error) {
    _, err := db.Exec("INSERT INTO users (auth_sub, display_name) VALUES (?, ?) ON DUPLICATE KEY UPDATE display_name=VALUES(display_name)", authSub, displayName)
    if err != nil { return 0, err }
    var id int64
    err = db.QueryRow("SELECT id FROM users WHERE auth_sub=?", authSub).Scan(&id)
    return id, err
}

func upsertMembership(db *sql.DB, tenantID, userID int64, role string) error {
    _, err := db.Exec("INSERT INTO tenant_memberships (tenant_id, user_id, role) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE role=VALUES(role)", tenantID, userID, role)
    return err
}

func ensureSamplePosts(db *sql.DB, tenantID, authorID int64) error {
    var cnt int
    if err := db.QueryRow("SELECT COUNT(*) FROM posts WHERE tenant_id=?", tenantID).Scan(&cnt); err != nil { return err }
    if cnt > 0 { return nil }
    // Insert 5 posts, each with 2 comments
    for i := 1; i <= 5; i++ {
        res, err := db.Exec("INSERT INTO posts (tenant_id, author_user_id, body, created_at) VALUES (?,?,?,?)",
            tenantID, authorID, fmt.Sprintf("hello world %d", i), time.Now())
        if err != nil { return err }
        postID, _ := res.LastInsertId()
        for j := 1; j <= 2; j++ {
            if _, err := db.Exec("INSERT INTO comments (tenant_id, post_id, author_user_id, body, created_at) VALUES (?,?,?,?,?)",
                tenantID, postID, authorID, fmt.Sprintf("comment %d-%d", i, j), time.Now()); err != nil { return err }
        }
    }
    return nil
}

func ensureSampleDM(db *sql.DB, tenantID, aliceID, bobID int64) error {
    // Create conversation
    var convID int64
    // Ensure only one conversation between alice and bob by checking existing members overlap of size 2
    row := db.QueryRow("SELECT c.id FROM conversations c JOIN conversation_members m1 ON m1.conversation_id=c.id AND m1.user_id=? JOIN conversation_members m2 ON m2.conversation_id=c.id AND m2.user_id=? WHERE c.tenant_id=? LIMIT 1", aliceID, bobID, tenantID)
    _ = row.Scan(&convID)
    if convID == 0 {
        res, err := db.Exec("INSERT INTO conversations (tenant_id, kind) VALUES (?, 'dm')", tenantID)
        if err != nil { return err }
        convID, _ = res.LastInsertId()
        if _, err := db.Exec("INSERT INTO conversation_members (conversation_id, user_id) VALUES (?, ?), (?, ?)", convID, aliceID, convID, bobID); err != nil { return err }
    }
    // Seed messages if none
    var msgCnt int
    if err := db.QueryRow("SELECT COUNT(*) FROM messages WHERE conversation_id=?", convID).Scan(&msgCnt); err != nil { return err }
    if msgCnt == 0 {
        msgs := []struct{ from int64; body string }{
            {aliceID, "Hey Bob"}, {bobID, "Hey Alice"}, {aliceID, "How are you?"},
        }
        for _, m := range msgs {
            if _, err := db.Exec("INSERT INTO messages (tenant_id, conversation_id, sender_user_id, body, created_at) VALUES (?,?,?,?,?)",
                tenantID, convID, m.from, m.body, time.Now()); err != nil { return err }
        }
    }
    return nil
}

func main() {
    db := openDB()
    defer db.Close()

    acmeID, err := upsertTenant(db, "acme", "Acme Inc")
    must(err, "upsert tenant acme")
    betaID, err := upsertTenant(db, "beta", "Beta LLC")
    must(err, "upsert tenant beta")

    _ = upsertTenantDomain(db, acmeID, "acme.localhost")
    _ = upsertTenantDomain(db, betaID, "beta.localhost")

    aliceID, err := upsertUser(db, "u_alice", "Alice")
    must(err, "upsert alice")
    bobID, err := upsertUser(db, "u_bob", "Bob")
    must(err, "upsert bob")
    caroID, err := upsertUser(db, "u_caro", "Caro")
    must(err, "upsert caro")

    must(upsertMembership(db, acmeID, aliceID, "owner"), "membership alice")
    must(upsertMembership(db, acmeID, bobID, "admin"), "membership bob")
    must(upsertMembership(db, acmeID, caroID, "member"), "membership caro")

    must(ensureSamplePosts(db, acmeID, aliceID), "seed posts")
    must(ensureSampleDM(db, acmeID, aliceID, bobID), "seed dm")

    log.Println("seed completed")
}
