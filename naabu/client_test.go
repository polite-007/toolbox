package naabu

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	naabuport "github.com/projectdiscovery/naabu/v2/pkg/port"
	"github.com/projectdiscovery/naabu/v2/pkg/protocol"
	naaburesult "github.com/projectdiscovery/naabu/v2/pkg/result"
	"github.com/projectdiscovery/naabu/v2/pkg/result/confidence"
)

func TestNaabuClientValidate(t *testing.T) {
	client := NewNaabuClient()

	if err := client.validate(nil, func(Result) {}); err == nil {
		t.Fatal("expected empty targets and empty targets file to fail")
	}

	if err := client.validate([]string{"scanme.nmap.org"}, nil); err == nil {
		t.Fatal("expected nil callback to fail")
	}

	client.config.ScanType = "invalid"
	if err := client.validate([]string{"scanme.nmap.org"}, func(Result) {}); err == nil {
		t.Fatal("expected invalid scan type to fail")
	}

	client = NewNaabuClient(WithTargetsFile("targets.txt"))
	if err := client.validate(nil, func(Result) {}); err != nil {
		t.Fatalf("expected targets file to satisfy target validation: %v", err)
	}
}

func TestToRunnerOptionsPortsAndTopPorts(t *testing.T) {
	// 未显式指定 Ports 时，默认走 TopPorts=100
	client := NewNaabuClient()
	options := client.toRunnerOptions([]string{"scanme.nmap.org"}, func(*naaburesult.HostResult) {})
	if options.Ports != "" || options.TopPorts != "100" {
		t.Fatalf("unexpected defaults: ports=%q topPorts=%q", options.Ports, options.TopPorts)
	}

	// 显式指定 Ports 时，清空 TopPorts，避免上游取并集导致多扫
	client = NewNaabuClient(WithPorts("80,443"))
	options = client.toRunnerOptions([]string{"scanme.nmap.org"}, func(*naaburesult.HostResult) {})
	if options.Ports != "80,443" || options.TopPorts != "" {
		t.Fatalf("unexpected explicit ports: ports=%q topPorts=%q", options.Ports, options.TopPorts)
	}
}

func TestToRunnerOptions(t *testing.T) {
	client := NewNaabuClient(
		WithTargetsFile("targets.txt"),
		WithPorts("80,443,u:53"),
		WithSynScan(),
		WithRate(100),
		WithRetries(2),
		WithTimeout(2*time.Second),
		WithThreads(10),
		WithWarmUpTime(1),
		WithHostPortConcurrency(3),
		WithVerify(),
		WithHostDiscovery(),
		WithPassive(),
		WithServiceDiscovery(),
		WithServiceVersion(),
		WithExcludeCDN(),
		WithProxy("127.0.0.1:1080", "user:pass"),
		WithResolvers("8.8.8.8,1.1.1.1"),
		WithSystemResolver(),
		WithSilent(false),
	)

	options := client.toRunnerOptions([]string{"scanme.nmap.org"}, func(*naaburesult.HostResult) {})
	if got, want := options.Host, []string{"scanme.nmap.org"}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("unexpected hosts: %#v", got)
	}
	if options.HostsFile != "targets.txt" {
		t.Fatalf("unexpected targets file: %s", options.HostsFile)
	}
	if options.Ports != "80,443,u:53" || options.TopPorts != "" {
		t.Fatalf("unexpected port options: ports=%q topPorts=%q", options.Ports, options.TopPorts)
	}
	if options.ScanType != "s" || options.Rate != 100 || options.Retries != 2 || options.Timeout != 2*time.Second || options.Threads != 10 || options.WarmUpTime != 1 || client.config.HostPortConcurrency != 3 {
		t.Fatalf("unexpected scan options: %#v", options)
	}
	if !options.Verify || !options.WithHostDiscovery || !options.Passive || !options.ServiceDiscovery || !options.ServiceVersion || !options.ExcludeCDN {
		t.Fatalf("expected feature flags to be enabled: %#v", options)
	}
	if options.Proxy != "127.0.0.1:1080" || options.ProxyAuth != "user:pass" || options.Resolvers != "8.8.8.8,1.1.1.1" || !options.SystemResolver {
		t.Fatalf("unexpected network options: %#v", options)
	}
	if options.Silent || !options.DisableStdout || options.OnResult == nil {
		t.Fatalf("unexpected output/callback options: %#v", options)
	}
}

func TestConvertResult(t *testing.T) {
	hostResult := &naaburesult.HostResult{
		Host:       "example.com",
		IP:         "93.184.216.34",
		Confidence: confidence.Low,
		MacAddress: "00:11:22:33:44:55",
		OS: &naaburesult.OSFingerprint{
			Target:     "93.184.216.34",
			DeviceType: "general purpose",
			Running:    "linux",
			OSCPE:      "cpe:/o:linux:linux_kernel",
			OSDetails:  "Linux",
		},
		Ports: []*naabuport.Port{
			{
				Port:     443,
				Protocol: protocol.TCP,
				TLS:      true,
				Service: &naabuport.Service{
					Name:       "https",
					Product:    "nginx",
					Version:    "1.24",
					Confidence: 10,
					CPEs:       []string{"cpe:/a:nginx:nginx:1.24"},
				},
			},
		},
	}

	results := toResults(hostResult)
	if len(results) != 1 {
		t.Fatalf("expected one port result, got %d", len(results))
	}
	if results[0].Host != "example.com" || results[0].IP != "93.184.216.34" || results[0].Port != 443 || results[0].Protocol != "tcp" || !results[0].TLS {
		t.Fatalf("unexpected port result: %#v", results[0])
	}
	if results[0].Service == nil || results[0].Service.Name != "https" || results[0].Service.Product != "nginx" || results[0].Service.Version != "1.24" {
		t.Fatalf("unexpected service result: %#v", results[0].Service)
	}

	converted := toHostResult(hostResult)
	if converted.Host != "example.com" || converted.IP != "93.184.216.34" || converted.Confidence != "low" || converted.MacAddress != "00:11:22:33:44:55" {
		t.Fatalf("unexpected host result: %#v", converted)
	}
	if converted.OS == nil || converted.OS.Running != "linux" {
		t.Fatalf("unexpected os result: %#v", converted.OS)
	}
	if len(converted.Ports) != 1 || converted.Ports[0].Service == nil || converted.Ports[0].Service.CPEs[0] != "cpe:/a:nginx:nginx:1.24" {
		t.Fatalf("unexpected converted ports: %#v", converted.Ports)
	}
}

func TestParseHostPortTarget(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		wantHost  string
		wantPort  int
		wantProto string
		wantErr   bool
	}{
		{name: "tcp default", raw: "1.1.1.1:443", wantHost: "1.1.1.1", wantPort: 443, wantProto: "tcp"},
		{name: "udp short prefix", raw: "u:8.8.8.8:53", wantHost: "8.8.8.8", wantPort: 53, wantProto: "udp"},
		{name: "udp prefix", raw: "udp:example.com:161", wantHost: "example.com", wantPort: 161, wantProto: "udp"},
		{name: "tcp prefix", raw: "tcp:example.com:8443", wantHost: "example.com", wantPort: 8443, wantProto: "tcp"},
		{name: "ipv6 unsupported", raw: "[2606:4700:4700::1111]:443", wantErr: true},
		{name: "missing port", raw: "example.com", wantErr: true},
		{name: "bad port", raw: "example.com:abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHostPortTarget(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Host != tt.wantHost || got.Port != tt.wantPort || got.Protocol != tt.wantProto {
				t.Fatalf("unexpected target: %#v", got)
			}
		})
	}
}

func TestGroupHostPortTargetsPreservesOrderAndDedupes(t *testing.T) {
	targets := []hostPortTarget{
		{Host: "1.1.1.1", Port: 443, Protocol: "tcp"},
		{Host: "1.1.1.1", Port: 80, Protocol: "tcp"},
		{Host: "1.1.1.1", Port: 443, Protocol: "tcp"},
		{Host: "8.8.8.8", Port: 53, Protocol: "udp"},
	}

	groups := groupHostPortTargets(targets)
	if got := hostPortExpression(groups["1.1.1.1"]); got != "443,80" {
		t.Fatalf("unexpected tcp expression: %s", got)
	}
	if got := hostPortExpression(groups["8.8.8.8"]); got != "u:53" {
		t.Fatalf("unexpected udp expression: %s", got)
	}
}

func TestOrderedHostPortResultsPreservesInputOrder(t *testing.T) {
	targets := []hostPortTarget{
		{Host: "1.1.1.1", Port: 443, Protocol: "tcp"},
		{Host: "8.8.8.8", Port: 53, Protocol: "udp"},
		{Host: "1.1.1.1", Port: 80, Protocol: "tcp"},
	}
	found := map[string]Result{
		hostPortResultKey("1.1.1.1", "tcp", 80):  {Host: "1.1.1.1", IP: "1.1.1.1", Port: 80, Protocol: "tcp"},
		hostPortResultKey("1.1.1.1", "tcp", 443): {Host: "1.1.1.1", IP: "1.1.1.1", Port: 443, Protocol: "tcp"},
		hostPortResultKey("8.8.8.8", "udp", 53):  {Host: "8.8.8.8", IP: "8.8.8.8", Port: 53, Protocol: "udp"},
	}

	results := orderedHostPortResults(targets, found)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0].Port != 443 || results[1].Port != 53 || results[2].Port != 80 {
		t.Fatalf("results are not in input order: %#v", results)
	}
}

func TestRunHostPortsValidation(t *testing.T) {
	client := NewNaabuClient()
	if err := client.RunHostPorts(context.Background(), nil, func(Result) {}); err == nil {
		t.Fatal("expected empty host ports to fail")
	}
	if err := client.RunHostPorts(context.Background(), []string{"1.1.1.1:443"}, nil); err == nil {
		t.Fatal("expected nil callback to fail")
	}
	if err := client.RunHostPorts(context.Background(), []string{"[2606:4700:4700::1111]:443"}, func(Result) {}); err == nil {
		t.Fatal("expected ipv6 host port to fail")
	}
}

func TestRunScanmeNmap(t *testing.T) {
	client := NewNaabuClient(
		WithPorts("80,443,22"),
		WithRate(100),
		WithTimeout(3*time.Second),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var gotResult bool
	err := client.Run(ctx, []string{"scanme.nmap.org"}, func(r Result) {
		gotResult = true
	})
	if err != nil {
		if isNetworkLikeError(err) || ctx.Err() != nil {
			t.Skipf("skipping real network scan test: %v", err)
		}
		t.Fatalf("run failed: %v", err)
	}
	if !gotResult {
		t.Log("scan completed without open ports; target state may have changed")
	}
}

func isNetworkLikeError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	for _, keyword := range []string{"timeout", "deadline", "no such host", "network", "connection", "dns", "resolver"} {
		if strings.Contains(message, keyword) {
			return true
		}
	}
	return false
}

func TestRunHostport(t *testing.T) {
	n := NewNaabuClient(
		WithPorts("-"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := n.Run(ctx, []string{"45.136.15.66"}, func(r Result) {
		fmt.Println(r)
		if r.Service != nil {
			fmt.Println(r.Service.Name)
		}
	}); err != nil {
		if isNetworkLikeError(err) || ctx.Err() != nil {
			t.Skipf("skipping real network scan test: %v", err)
		}
		t.Fatalf("run failed: %v", err)
	}
}
