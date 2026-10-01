package auth

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// testStore migrates the disposable database named by AUTH_TEST_DATABASE_URL
// (skipped when unset; never point it at a working database).
func testStore(t *testing.T) *PGX {
	t.Helper()
	dsn := os.Getenv("AUTH_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("AUTH_TEST_DATABASE_URL not set: skipping database test")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	files, _ := filepath.Glob("../../cmd/goCloudAuthServer/db/migrations/*.up.sql")
	sort.Strings(files)
	for _, f := range files {
		sql, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			t.Fatalf("migrate %s: %v", f, err)
		}
	}
	if _, err := pool.Exec(ctx, "TRUNCATE go_auth.users CASCADE"); err != nil {
		t.Fatalf("reset users: %v", err)
	}
	return &PGX{Conn: pool, log: slog.Default()}
}

// TestUpsertLinksOnlyVerifiedEmail covers account linking by e-mail: a login
// of another provider reuses an account only when its e-mail is verified;
// otherwise it is refused (ErrEmailInUse) and the account keeps its provider.
func TestUpsertLinksOnlyVerifiedEmail(t *testing.T) {
	db := testStore(t)
	ctx := context.Background()
	google := OAuthUserInfo{Provider: "google", ProviderID: "g-1", Email: "ada@example.org", Name: "Ada", EmailVerified: true}
	owner, err := db.UpsertByProvider(ctx, google)
	if err != nil {
		t.Fatalf("first login: %v", err)
	}

	intruder := OAuthUserInfo{Provider: "microsoft", ProviderID: "m-1", Email: "ada@example.org", Name: "Not Ada"}
	if _, err := db.UpsertByProvider(ctx, intruder); !errors.Is(err, ErrEmailInUse) {
		t.Fatalf("an unverified e-mail of another account: want ErrEmailInUse, got %v", err)
	}
	again, err := db.UpsertByProvider(ctx, google)
	if err != nil || again.ID != owner.ID || again.Provider != "google" {
		t.Fatalf("the owner keeps its account and provider: %+v (%v)", again, err)
	}

	github := OAuthUserInfo{Provider: "github", ProviderID: "h-1", Email: "ada@example.org", Name: "Ada", EmailVerified: true}
	linked, err := db.UpsertByProvider(ctx, github)
	if err != nil || linked.ID != owner.ID {
		t.Fatalf("a verified e-mail links the same account: %+v (%v)", linked, err)
	}

	newcomer := OAuthUserInfo{Provider: "microsoft", ProviderID: "m-2", Email: "bob@example.org", Name: "Bob"}
	created, err := db.UpsertByProvider(ctx, newcomer)
	if err != nil || created.ID == owner.ID {
		t.Fatalf("a new e-mail creates an account, verified or not: %+v (%v)", created, err)
	}
	if back, err := db.UpsertByProvider(ctx, newcomer); err != nil || back.ID != created.ID {
		t.Fatalf("the same provider identity finds its account: %+v (%v)", back, err)
	}
}
