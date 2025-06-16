// internal/vault/vault.go
//
// Vault client wrapper for Adept.
//
// Context
// -------
//   - Provides a concurrency‑safe singleton around the HashiCorp Vault Go SDK.
//   - Adds background token renewal, simple KV‑v2 helpers, and per‑key caching.
//   - Follows the "Adept Comment‑Style Guide"—header block, section underlines,
//     Oxford commas, two spaces after periods, no m‑dash.
//
// Public workflow
// ---------------
//  1. cli, err := vault.New(ctx, log.Printf)       // during boot.
//  2. pw,  err := cli.GetKV(ctx, path, key, ttl)   // anywhere in the app.
//
// Build tags: none.
package vault

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	vault "github.com/hashicorp/vault/api"
)

//
// SECTION 1.  Public façade
//

// Client is safe for concurrent use.  Create once at startup and inject it via
// your DI container.  Zero value is invalid.
type Client struct {
	api   *vault.Client
	logFn func(string, ...any)

	cacheMu sync.RWMutex
	cache   map[string]cached // canonical path#key → value + expiry.
}

type cached struct {
	val string
	exp time.Time
}

// New constructs a Vault client and starts a background token‑renewal loop.
//
// Environment expectations
// ------------------------
// • VAULT_ADDR   – scheme and host of the Vault server.
// • VAULT_TOKEN  – initial token (falls back to ~/.vault‑token).
func New(ctx context.Context, logFn func(string, ...any)) (*Client, error) {
	if logFn == nil {
		logFn = func(string, ...any) {}
	}

	cfg := vault.DefaultConfig()
	if err := cfg.ReadEnvironment(); err != nil {
		return nil, fmt.Errorf("vault env cfg: %w", err)
	}

	apiCli, err := vault.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("vault api: %w", err)
	}

	if tok := os.Getenv("VAULT_TOKEN"); tok != "" {
		apiCli.SetToken(tok)
	}

	c := &Client{
		api:   apiCli,
		logFn: logFn,
		cache: make(map[string]cached),
	}

	go c.renewLoop(ctx)

	return c, nil
}

// GetKV fetches a single key from a KV‑v2 secret.  If ttl > 0 the result is
// cached for that duration.  Subsequent callers within the TTL receive the
// cached copy.
func (c *Client) GetKV(ctx context.Context, secretPath, key string, ttl time.Duration) (string, error) {
	if secretPath == "" || key == "" {
		return "", errors.New("secret path and key must be non‑empty")
	}

	canonical := secretPath + "#" + key

	if ttl > 0 {
		c.cacheMu.RLock()
		if cv, ok := c.cache[canonical]; ok && time.Now().Before(cv.exp) {
			c.cacheMu.RUnlock()
			return cv.val, nil
		}
		c.cacheMu.RUnlock()
	}

	mount, rel := splitMount(secretPath)
	sec, err := c.api.KVv2(mount).Get(ctx, rel)
	if err != nil {
		return "", fmt.Errorf("vault get %s: %w", secretPath, err)
	}

	raw, ok := sec.Data[key]
	if !ok {
		return "", fmt.Errorf("key %q not found in secret %q", key, secretPath)
	}

	sval, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("value at %s#%s is not a string", secretPath, key)
	}

	if ttl > 0 {
		c.cacheMu.Lock()
		c.cache[canonical] = cached{val: sval, exp: time.Now().Add(ttl)}
		c.cacheMu.Unlock()
	}

	return sval, nil
}

//
// SECTION 2.  Background token renewal
//

func (c *Client) renewLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Probe the current token.
		sec, err := c.api.Auth().Token().RenewSelf(0)
		if err != nil {
			c.logFn("vault: token renew self failed: %v", err)
			backoff(ctx, 30*time.Second)
			continue
		}

		if sec == nil || !sec.Auth.Renewable {
			c.logFn("vault: token is not renewable – sleeping 1h")
			backoff(ctx, time.Hour)
			continue
		}

		renewer, err := c.api.NewRenewer(&vault.RenewerInput{
			Secret: sec,
			Grace:  15 * time.Second,
		})
		if err != nil {
			c.logFn("vault: renewer init error: %v", err)
			backoff(ctx, 30*time.Second)
			continue
		}

		go renewer.Renew()

		for {
			select {
			case <-ctx.Done():
				renewer.Stop()
				return
			case err := <-renewer.DoneCh():
				renewer.Stop()
				if err != nil {
					c.logFn("vault: token renewal stopped: %v", err)
				}
				backoff(ctx, 15*time.Second)
				goto probe
			case ev := <-renewer.RenewCh():
				if ev != nil && ev.Secret != nil && ev.Secret.Auth != nil {
					c.logFn("vault: token renewed, ttl=%ds", ev.Secret.Auth.LeaseDuration)
				}
			}
		}
	probe:
	}
}

//
// SECTION 3.  Helpers
//

func splitMount(p string) (mount, rel string) {
	if p == "" {
		return "", ""
	}
	parts := strings.SplitN(p, "/", 2)
	mount = parts[0]
	if len(parts) == 2 {
		rel = parts[1]
	}
	return
}

func backoff(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}
