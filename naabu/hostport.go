package naabu

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type hostPortTarget struct {
	Host     string
	Port     int
	Protocol string
}

func parseHostPortTarget(raw string) (hostPortTarget, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return hostPortTarget{}, fmt.Errorf("empty host port target")
	}

	protocol := "tcp"
	if rest, ok := strings.CutPrefix(value, "u:"); ok {
		protocol = "udp"
		value = rest
	} else if rest, ok := strings.CutPrefix(value, "udp:"); ok {
		protocol = "udp"
		value = rest
	} else if rest, ok := strings.CutPrefix(value, "tcp:"); ok {
		value = rest
	}

	if strings.Contains(value, "[") || strings.Contains(value, "]") || strings.Count(value, ":") != 1 {
		return hostPortTarget{}, fmt.Errorf("invalid host port target %q: ipv6 is not supported", raw)
	}

	host, portValue, ok := strings.Cut(value, ":")
	if !ok || host == "" || portValue == "" {
		return hostPortTarget{}, fmt.Errorf("invalid host port target %q", raw)
	}

	portNumber, err := strconv.Atoi(portValue)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return hostPortTarget{}, fmt.Errorf("invalid port in host port target %q", raw)
	}

	return hostPortTarget{Host: host, Port: portNumber, Protocol: protocol}, nil
}

func groupHostPortTargets(targets []hostPortTarget) map[string][]hostPortTarget {
	groups := make(map[string][]hostPortTarget)
	seen := make(map[string]map[string]struct{})
	for _, target := range targets {
		key := hostPortKey(target)
		if seen[target.Host] == nil {
			seen[target.Host] = make(map[string]struct{})
		}
		if _, ok := seen[target.Host][key]; ok {
			continue
		}
		seen[target.Host][key] = struct{}{}
		groups[target.Host] = append(groups[target.Host], target)
	}
	return groups
}

func hostPortKey(target hostPortTarget) string {
	return fmt.Sprintf("%s:%d", target.Protocol, target.Port)
}

func hostPortExpression(targets []hostPortTarget) string {
	ports := make([]string, 0, len(targets))
	for _, target := range targets {
		if target.Protocol == "udp" {
			ports = append(ports, fmt.Sprintf("u:%d", target.Port))
			continue
		}
		ports = append(ports, strconv.Itoa(target.Port))
	}
	return strings.Join(ports, ",")
}

func cloneConfig(cfg *Config) *Config {
	cloned := *cfg
	return &cloned
}

func orderedHostPortResults(targets []hostPortTarget, found map[string]Result) []Result {
	results := make([]Result, 0, len(targets))
	for _, target := range targets {
		if result, ok := found[hostPortResultKey(target.Host, target.Protocol, target.Port)]; ok {
			results = append(results, result)
		}
	}
	return results
}

func hostPortResultKey(host, protocol string, port int) string {
	return fmt.Sprintf("%s|%s:%d", host, protocol, port)
}

func collectHostPortResults(mu *sync.Mutex, found map[string]Result, inputHost string, result Result) {
	mu.Lock()
	defer mu.Unlock()
	found[hostPortResultKey(inputHost, result.Protocol, result.Port)] = result
	found[hostPortResultKey(result.Host, result.Protocol, result.Port)] = result
	if result.IP != "" {
		found[hostPortResultKey(result.IP, result.Protocol, result.Port)] = result
	}
}
