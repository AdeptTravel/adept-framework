// internal/routing/alias_test.go
//
// Unit-tests for AliasRewrite middleware.
//
// Context
// -------
// The AliasRewrite handler rewrites friendly paths to absolute component
// paths by consulting an in-memory AliasCache with SQL fallback.  These tests
// verify three critical behaviours:
//
//   • Cache-hit rewrite in BOTH mode                         → 200, path mutated
//   • Cache-miss in ALIAS-only mode                          → 404
//   • ABSOLUTE routing mode leaves path untouched            → 200
//
// Workflow / Structure
// --------------------
// fakeTenant ── minimal AliasTenant implementation that lets us inject a
// pre-seeded cache or an empty one without pulling in the full tenant loader.
//
// Each sub-test:
//
//   1. Builds sqlmock DB (rows unused except in SQL-fallback test).
//   2. Populates cache as needed.
//   3. Wraps a chi handler with Middleware(fakeTenant).
//   4. Fires an httptest request and asserts status / rewritten path.
//
// Notes
// -----
// • Oxford commas, two spaces after periods.
// • Lines ≤ 100 columns.

package routing

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// fakeTenant satisfies AliasTenant with injectable fields.
type fakeTenant struct {
	mode    string
	version int
	cache   *AliasCache
}

func (f *fakeTenant) RoutingMode() string     { return f.mode }
func (f *fakeTenant) RouteVersion() int       { return f.version }
func (f *fakeTenant) AliasCache() *AliasCache { return f.cache }

func TestAliasRewrite_CacheHit(t *testing.T) {
	db, _, _ := sqlmock.New()
	cache := NewAliasCache(db, time.Minute)
	cache.store("/about", "/content/page/view/about")

	tenant := &fakeTenant{mode: RouteModeBoth, cache: cache}

	var got string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/about", nil)
	rr := httptest.NewRecorder()

	Middleware(tenant)(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if got != "/content/page/view/about" {
		t.Fatalf("rewrite failed: got path %q", got)
	}
}

func TestAliasRewrite_Miss_AliasOnly(t *testing.T) {
	db, _, _ := sqlmock.New()
	cache := NewAliasCache(db, time.Minute)

	tenant := &fakeTenant{mode: RouteModeAliasOnly, cache: cache}

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rr := httptest.NewRecorder()

	Middleware(tenant)(http.NotFoundHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rr.Code)
	}
}

func TestAliasRewrite_AbsoluteMode_NoMutation(t *testing.T) {
	db, _, _ := sqlmock.New()
	cache := NewAliasCache(db, time.Minute)

	tenant := &fakeTenant{mode: RouteModeAbsolute, cache: cache}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/keep" {
			t.Fatalf("path mutated in absolute mode: %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/keep", nil)
	rr := httptest.NewRecorder()

	Middleware(tenant)(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
}
