package ipinfo

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestLookup(t *testing.T) {
	client, err := NewIpinfoClient(WithTimeout(15 * time.Second))
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	info, err := client.Lookup(ctx, "8.8.8.8")
	if err != nil {
		if isNetworkLikeError(err) || ctx.Err() != nil {
			t.Skipf("skipping real network test: %v", err)
		}
		t.Fatalf("lookup failed: %v", err)
	}
	if info.IP != "8.8.8.8" {
		t.Fatalf("unexpected ip: %q", info.IP)
	}
	if info.Source == "" {
		t.Fatal("expected source to be set")
	}
	t.Logf("source=%s country=%q countryCode=%q org=%q asn=%q city=%q", info.Source, info.Country, info.CountryCode, info.Org, info.ASN, info.City)
}

func TestLookupInvalidIP(t *testing.T) {
	client, err := NewIpinfoClient()
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	for _, ip := range []string{"", "abc", "999.1.1.1", "8.8.8.8.8"} {
		if _, err := client.Lookup(context.Background(), ip); err == nil {
			t.Fatalf("expected invalid ip %q to fail", ip)
		}
	}
}

func TestLookupWithSource(t *testing.T) {
	client, err := NewIpinfoClient(WithTimeout(15*time.Second), WithSource("ipinfo.io"))
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	info, err := client.Lookup(ctx, "8.8.8.8")
	if err != nil {
		if isNetworkLikeError(err) || ctx.Err() != nil {
			t.Skipf("skipping real network test: %v", err)
		}
		t.Fatalf("lookup failed: %v", err)
	}
	if info.Source != "ipinfo.io" {
		t.Fatalf("unexpected source: %q", info.Source)
	}
}

func TestLookupUnknownSource(t *testing.T) {
	client, err := NewIpinfoClient(WithSource("nonexistent"))
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	if _, err := client.Lookup(context.Background(), "8.8.8.8"); err == nil {
		t.Fatal("expected unknown source to fail")
	}
}

func TestParseIpinfo(t *testing.T) {
	body := []byte(`{"ip":"8.8.8.8","hostname":"dns.google","city":"Mountain View","region":"California","country":"US","loc":"38.0088,-122.1175","org":"AS15169 Google LLC","postal":"94043","timezone":"America/Los_Angeles"}`)
	info, err := parseIpinfo(body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if info.IP != "8.8.8.8" || info.Hostname != "dns.google" || info.City != "Mountain View" || info.CountryCode != "US" || info.ASN != "AS15169" || info.Org != "AS15169 Google LLC" {
		t.Fatalf("unexpected info: %#v", info)
	}
	if info.Latitude != 38.0088 || info.Longitude != -122.1175 {
		t.Fatalf("unexpected loc: %v,%v", info.Latitude, info.Longitude)
	}
}

func TestParseIpwhois(t *testing.T) {
	body := []byte(`{"ip":"8.8.8.8","success":true,"continent":"North America","country":"United States","country_code":"US","region":"California","city":"San Jose","latitude":37.3361663,"longitude":-121.8905913,"postal":"95113","connection":{"asn":15169,"org":"Google LLC","isp":"Google LLC"},"timezone":{"id":"America/Los_Angeles"}}`)
	info, err := parseIpwhois(body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if info.Country != "United States" || info.CountryCode != "US" || info.Continent != "North America" || info.Org != "Google LLC" || info.ASN != "AS15169" || info.ISP != "Google LLC" {
		t.Fatalf("unexpected info: %#v", info)
	}
}

func TestParseIp9(t *testing.T) {
	body := []byte(`{"data":{"ip":"8.8.8.8","country":"美国","country_code":"us","prov":"","city":"","isp":"Google Cloud","lng":"-97.82","lat":"37.75"}}`)
	info, err := parseIp9(body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if info.Country != "美国" || info.CountryCode != "US" || info.ISP != "Google Cloud" || info.Latitude != 37.75 || info.Longitude != -97.82 {
		t.Fatalf("unexpected info: %#v", info)
	}
}

func TestParseIpdata(t *testing.T) {
	body := []byte(`{"ip":"8.8.8.8","success":true,"continent":"North America","country":"United States","country_code":"US","region":"California","city":"Mountain View","latitude":37.751,"longitude":-97.822,"postal":"94043","asn":15169,"asn_org":"Google LLC","isp":"Google LLC","timezone":"America/Chicago"}`)
	info, err := parseIpdata(body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if info.Country != "United States" || info.ASN != "AS15169" || info.Org != "Google LLC" || info.ISP != "Google LLC" || info.Timezone != "America/Chicago" {
		t.Fatalf("unexpected info: %#v", info)
	}
}

func isNetworkLikeError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	for _, keyword := range []string{"timeout", "deadline", "no such host", "network", "connection", "dns", "refused", "unreachable", "eof"} {
		if strings.Contains(message, keyword) {
			return true
		}
	}
	return false
}
