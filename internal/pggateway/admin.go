package pggateway

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"intra/internal/broker"
)

func RoleForGrant(id string) string {
	r := strings.NewReplacer("-", "_", ".", "_").Replace(strings.ToLower(id))
	return "intra_" + r
}

type PGKiller struct {
	Admin   *pgxpool.Pool
	Gateway *Gateway
	Objects []string
}

func randHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (k *PGKiller) OpenGrant(ctx context.Context, g broker.Grant) (string, error) {
	secret, err := randHex(24)
	if err != nil {
		return "", err
	}
	role := RoleForGrant(g.ID)
	if k.Admin == nil {
		return "", fmt.Errorf("postgres admin unavailable")
	}
	objects := k.Objects
	if len(objects) == 0 {
		objects = []string{"secrets"}
	}
	_, err = k.Admin.Exec(ctx, "SELECT intra_create_grant_role($1,$2,$3,$4)", role, secret, g.ExpiresAt.Format(time.RFC3339), objects)
	if err != nil {
		return "", fmt.Errorf("create grant role: %w", err)
	}
	return secret, nil
}

func (k *PGKiller) KillGrant(ctx context.Context, g broker.Grant) error {
	if k.Gateway != nil {
		k.Gateway.CloseGrant(g.ID)
	}
	if k.Admin == nil {
		return nil
	}
	role := RoleForGrant(g.ID)
	_, err := k.Admin.Exec(ctx, "SELECT intra_drop_grant_role($1)", role)
	return err
}

// ConnectAdmin opens the broker admin pool. The admin database role must have
// EXECUTE on the intra_* functions only. It holds no SELECT on protected data.
func ConnectAdmin(ctx context.Context, conn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(conn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 4
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func ParseBase(conn string) (*pgx.ConnConfig, error) {
	return pgx.ParseConfig(conn)
}
