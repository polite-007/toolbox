package naabu

import (
	"fmt"

	naabuport "github.com/projectdiscovery/naabu/v2/pkg/port"
	naaburesult "github.com/projectdiscovery/naabu/v2/pkg/result"
	"github.com/projectdiscovery/naabu/v2/pkg/result/confidence"
)

// Result 表示单个开放端口结果。
type Result struct {
	Host     string
	IP       string
	Port     int
	Protocol string
	TLS      bool
	Service  *Service
}

// HostResult 表示单个主机的聚合扫描结果。
type HostResult struct {
	Host       string
	IP         string
	Ports      []Port
	Confidence string
	OS         *OSFingerprint
	MacAddress string
}

// Port 表示一个开放端口。
type Port struct {
	Port     int
	Protocol string
	TLS      bool
	Service  *Service
}

// Service 表示端口上的服务识别信息。
type Service struct {
	DeviceType  string
	ExtraInfo   string
	HighVersion string
	Hostname    string
	LowVersion  string
	Method      string
	Name        string
	OSType      string
	Product     string
	Proto       string
	RPCNum      string
	ServiceFP   string
	Tunnel      string
	Version     string
	Confidence  int
	CPEs        []string
}

// OSFingerprint 表示主机 OS 指纹信息。
type OSFingerprint struct {
	Target     string
	DeviceType string
	Running    string
	OSCPE      string
	OSDetails  string
}

func toResults(hostResult *naaburesult.HostResult) []Result {
	if hostResult == nil {
		return nil
	}

	results := make([]Result, 0, len(hostResult.Ports))
	for _, p := range hostResult.Ports {
		if p == nil {
			continue
		}
		results = append(results, Result{
			Host:     hostResult.Host,
			IP:       hostResult.IP,
			Port:     p.Port,
			Protocol: p.Protocol.String(),
			TLS:      p.TLS,
			Service:  toService(p.Service),
		})
	}
	return results
}

func toHostResult(hostResult *naaburesult.HostResult) HostResult {
	if hostResult == nil {
		return HostResult{}
	}

	ports := make([]Port, 0, len(hostResult.Ports))
	for _, p := range hostResult.Ports {
		if p == nil {
			continue
		}
		ports = append(ports, toPort(p))
	}

	return HostResult{
		Host:       hostResult.Host,
		IP:         hostResult.IP,
		Ports:      ports,
		Confidence: confidenceString(hostResult.Confidence),
		OS:         toOSFingerprint(hostResult.OS),
		MacAddress: hostResult.MacAddress,
	}
}

func toPort(p *naabuport.Port) Port {
	return Port{
		Port:     p.Port,
		Protocol: p.Protocol.String(),
		TLS:      p.TLS,
		Service:  toService(p.Service),
	}
}

func toService(service *naabuport.Service) *Service {
	if service == nil {
		return nil
	}

	return &Service{
		DeviceType:  service.DeviceType,
		ExtraInfo:   service.ExtraInfo,
		HighVersion: service.HighVersion,
		Hostname:    service.Hostname,
		LowVersion:  service.LowVersion,
		Method:      service.Method,
		Name:        service.Name,
		OSType:      service.OSType,
		Product:     service.Product,
		Proto:       service.Proto,
		RPCNum:      service.RPCNum,
		ServiceFP:   service.ServiceFP,
		Tunnel:      service.Tunnel,
		Version:     service.Version,
		Confidence:  service.Confidence,
		CPEs:        append([]string(nil), service.CPEs...),
	}
}

func toOSFingerprint(os *naaburesult.OSFingerprint) *OSFingerprint {
	if os == nil {
		return nil
	}
	return &OSFingerprint{
		Target:     os.Target,
		DeviceType: os.DeviceType,
		Running:    os.Running,
		OSCPE:      os.OSCPE,
		OSDetails:  os.OSDetails,
	}
}

func confidenceString(level confidence.ConfidenceLevel) string {
	switch level {
	case confidence.Normal:
		return "normal"
	case confidence.Low:
		return "low"
	default:
		return fmt.Sprintf("unknown(%d)", level)
	}
}
