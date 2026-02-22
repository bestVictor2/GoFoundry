package geecache

import (
	"errors"
	"testing"
	"time"
)

func TestGroupTTLAndRemove(t *testing.T) {
	loadCount := 0
	g := NewGroupWithOptions("ttl-test", GetterFunc(func(key string) ([]byte, error) {
		loadCount++
		return []byte("value-" + key), nil
	}), 2<<10, WithTTL(40*time.Millisecond))

	if _, err := g.Get("tom"); err != nil {
		t.Fatalf("first get failed: %v", err)
	}
	if _, err := g.Get("tom"); err != nil {
		t.Fatalf("second get failed: %v", err)
	}
	if loadCount != 1 {
		t.Fatalf("expected load count 1 before ttl expiration, got %d", loadCount)
	}

	time.Sleep(70 * time.Millisecond)
	if _, err := g.Get("tom"); err != nil {
		t.Fatalf("third get after ttl failed: %v", err)
	}
	if loadCount != 2 {
		t.Fatalf("expected load count 2 after ttl expiration, got %d", loadCount)
	}

	g.Remove("tom")
	if _, err := g.Get("tom"); err != nil {
		t.Fatalf("get after remove failed: %v", err)
	}
	if loadCount != 3 {
		t.Fatalf("expected load count 3 after remove, got %d", loadCount)
	}

	stats := g.Stats()
	if stats.Hits != 1 || stats.Misses != 3 || stats.Loads != 3 || stats.LocalLoads != 3 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}

type mockPeerGetter struct {
	calls int
}

func (m *mockPeerGetter) Get(_ string, _ string) ([]byte, error) {
	m.calls++
	return []byte("peer-value"), nil
}

type mockPeerPicker struct {
	peer PeerGetter
}

func (m *mockPeerPicker) PickPeer(_ string) (PeerGetter, bool) {
	return m.peer, true
}

func TestPeerLoadPopulatesLocalCache(t *testing.T) {
	peer := &mockPeerGetter{}
	g := NewGroupWithOptions("peer-fill-test", GetterFunc(func(key string) ([]byte, error) {
		return nil, errors.New("local should not be called")
	}), 2<<10)
	g.RegisterPeers(&mockPeerPicker{peer: peer})

	view, err := g.Get("sam")
	if err != nil {
		t.Fatalf("first peer get failed: %v", err)
	}
	if got := view.String(); got != "peer-value" {
		t.Fatalf("unexpected peer value: %s", got)
	}

	view, err = g.Get("sam")
	if err != nil {
		t.Fatalf("second get failed: %v", err)
	}
	if got := view.String(); got != "peer-value" {
		t.Fatalf("unexpected cached value: %s", got)
	}
	if peer.calls != 1 {
		t.Fatalf("expected peer called once, got %d", peer.calls)
	}

	stats := g.Stats()
	if stats.PeerLoads != 1 || stats.Hits != 1 || stats.Misses != 1 {
		t.Fatalf("unexpected peer stats: %+v", stats)
	}
}

func TestGroupGetManyAndRemoveMany(t *testing.T) {
	src := map[string]string{
		"a": "1",
		"b": "2",
	}
	g := NewGroupWithOptions("batch-test", GetterFunc(func(key string) ([]byte, error) {
		if v, ok := src[key]; ok {
			return []byte(v), nil
		}
		return nil, errors.New("not found")
	}), 2<<10)

	values, errs := g.GetMany("a", "b", "x")
	if len(values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(values))
	}
	if values["a"].String() != "1" || values["b"].String() != "2" {
		t.Fatalf("unexpected batch values: %+v", values)
	}
	if errs == nil || errs["x"] == nil {
		t.Fatalf("expected error for key x")
	}

	stats := g.Stats()
	if stats.Entries != 2 {
		t.Fatalf("expected entries=2, got %d", stats.Entries)
	}
	if stats.CacheBytes <= 0 {
		t.Fatalf("expected positive cache bytes, got %d", stats.CacheBytes)
	}

	g.RemoveMany("a", "b")
	stats = g.Stats()
	if stats.Entries != 0 {
		t.Fatalf("expected entries=0 after remove many, got %d", stats.Entries)
	}
	if stats.Evictions < 2 {
		t.Fatalf("expected at least 2 evictions, got %d", stats.Evictions)
	}
}
