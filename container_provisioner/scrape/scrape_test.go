package scrape

import (
	"context"
	"net/netip"
	"testing"
	"time"

	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/containers"
	"github.com/algo7/TripAdvisor-Review-Scraper/container_provisioner/internal/redistest"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
)

// fakeDocker lists a fixed set of containers. Every other Docker call panics, because the embedded interface is nil.
type fakeDocker struct {
	containers.ContainerClient
	items []container.Summary
}

func (f *fakeDocker) ContainerList(_ context.Context, _ client.ContainerListOptions) (client.ContainerListResult, error) {
	return client.ContainerListResult{Items: f.items}, nil
}

// testProxy is a proxy container for one test. Its ID is random, so lock keys never clash between tests.
type testProxy struct {
	id        string
	ip        string
	socksPort string
}

func newTestProxy(t *testing.T, ip string, socksPort string) testProxy {
	return testProxy{id: redistest.UniqueID(t), ip: ip, socksPort: socksPort}
}

// summary describes the proxy the way Docker lists it. An empty ip leaves it off the scraper_vpn network.
func (p testProxy) summary() container.Summary {
	summary := container.Summary{
		ID: p.id,
		Labels: map[string]string{
			"TaskOwner":        "PROXY",
			"vpn.region":       "CH",
			"proxy.socks.port": p.socksPort,
		},
		NetworkSettings: &container.NetworkSettingsSummary{Networks: map[string]*network.EndpointSettings{}},
	}
	if p.ip != "" {
		summary.NetworkSettings.Networks["scraper_vpn"] = &network.EndpointSettings{IPAddress: netip.MustParseAddr(p.ip)}
	}
	return summary
}

// containerID is the shortened ID the provisioner uses for the proxy.
func (p testProxy) containerID() string {
	return p.id[:12]
}

func newTestScraper(t *testing.T, proxies ...testProxy) *Scraper {
	t.Helper()
	r := redistest.New(t)

	items := []container.Summary{}
	for _, p := range proxies {
		items = append(items, p.summary())
		redistest.Cleanup(t, r, proxyLockKey(p.containerID()))
	}

	cm := &containers.ContainerManager{Client: &fakeDocker{items: items}}
	return NewScraper(cm, nil, r)
}

// setLockTiming shortens the proxy lock timing so expiry can be tested in real time.
func setLockTiming(t *testing.T, ttl time.Duration, renewInterval time.Duration) {
	previousTTL, previousInterval := proxyLockTTL, proxyLockRenewInterval
	proxyLockTTL, proxyLockRenewInterval = ttl, renewInterval
	t.Cleanup(func() { proxyLockTTL, proxyLockRenewInterval = previousTTL, previousInterval })
}

func (s *Scraper) lockOwner(t *testing.T, proxyID string) string {
	return s.Redis.Client.Get(t.Context(), proxyLockKey(proxyID)).Val()
}

func TestAcquireProxyContainerHandsOutEachProxyOnce(t *testing.T) {
	a := newTestProxy(t, "172.28.2.2", "8881")
	b := newTestProxy(t, "172.28.2.3", "8881")
	s := newTestScraper(t, a, b)

	first, err := s.AcquireProxyContainer()
	if err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	if first.ContainerID != a.containerID() || first.ProxyAddress != "socks5://172.28.2.2:8881" {
		t.Fatalf("first acquire = %s %s, want %s socks5://172.28.2.2:8881", first.ContainerID, first.ProxyAddress, a.containerID())
	}

	second, err := s.AcquireProxyContainer()
	if err != nil || second.ContainerID != b.containerID() {
		t.Fatalf("second acquire = %q, %v, want %s", second.ContainerID, err, b.containerID())
	}

	if _, err := s.AcquireProxyContainer(); err == nil {
		t.Fatal("third acquire succeeded, want an error with both proxies in use")
	}

	s.ReleaseProxyContainer(first)
	again, err := s.AcquireProxyContainer()
	if err != nil || again.ContainerID != a.containerID() {
		t.Fatalf("acquire after release = %q, %v, want %s", again.ContainerID, err, a.containerID())
	}
}

func TestAcquireProxyContainerSkipsUnusableProxies(t *testing.T) {
	offNetwork := newTestProxy(t, "", "8881")
	noPort := newTestProxy(t, "172.28.2.3", "")
	usable := newTestProxy(t, "172.28.2.4", "8881")
	s := newTestScraper(t, offNetwork, noPort, usable)

	proxy, err := s.AcquireProxyContainer()
	if err != nil || proxy.ContainerID != usable.containerID() {
		t.Fatalf("acquire = %q, %v, want %s", proxy.ContainerID, err, usable.containerID())
	}
	for _, p := range []testProxy{offNetwork, noPort} {
		if s.lockOwner(t, p.containerID()) != "" {
			t.Errorf("unusable proxy %s was left locked", p.containerID())
		}
	}
}

func TestHoldProxyContainerKeepsLockPastTTL(t *testing.T) {
	setLockTiming(t, 300*time.Millisecond, 30*time.Millisecond)
	p := newTestProxy(t, "172.28.2.2", "8881")
	s := newTestScraper(t, p)

	proxy, err := s.AcquireProxyContainer()
	if err != nil {
		t.Fatal(err)
	}
	go s.HoldProxyContainer(t.Context(), proxy)

	// Several lock lifetimes pass. Without renewals the lock would have expired.
	time.Sleep(4 * proxyLockTTL)

	if owner := s.lockOwner(t, proxy.ContainerID); owner != proxy.lockToken {
		t.Fatalf("lock owner = %q, want the holding scrape's token", owner)
	}
	if _, err := s.AcquireProxyContainer(); err == nil {
		t.Fatal("another scrape got the proxy while it was held")
	}
}

func TestReleaseProxyContainerLeavesNewOwnerAlone(t *testing.T) {
	setLockTiming(t, 100*time.Millisecond, time.Hour)
	p := newTestProxy(t, "172.28.2.2", "8881")
	s := newTestScraper(t, p)

	stale, err := s.AcquireProxyContainer()
	if err != nil {
		t.Fatal(err)
	}

	// The first scrape stopped renewing, so its lock expires and a second scrape takes the proxy
	time.Sleep(2 * proxyLockTTL)
	current, err := s.AcquireProxyContainer()
	if err != nil {
		t.Fatalf("acquire after expiry: %v", err)
	}

	s.ReleaseProxyContainer(stale)

	if owner := s.lockOwner(t, current.ContainerID); owner != current.lockToken {
		t.Fatalf("lock owner after stale release = %q, want the second scrape's token", owner)
	}
}

func TestHoldProxyContainerStopsWhenLockIsLost(t *testing.T) {
	setLockTiming(t, 100*time.Millisecond, 20*time.Millisecond)
	p := newTestProxy(t, "172.28.2.2", "8881")
	s := newTestScraper(t, p)

	stale, err := s.AcquireProxyContainer()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * proxyLockTTL)
	current, err := s.AcquireProxyContainer()
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	go func() {
		s.HoldProxyContainer(t.Context(), stale)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("hold kept running after losing the lock")
	}
	if owner := s.lockOwner(t, current.ContainerID); owner != current.lockToken {
		t.Fatal("the stale hold changed the new owner's lock")
	}
}
