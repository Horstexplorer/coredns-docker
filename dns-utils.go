package docker

import (
	"fmt"
	"net"
	"strings"

	"github.com/miekg/dns"
)

func NewDNSResponse(request *dns.Msg, authoritative bool, records []dns.RR) *dns.Msg {
	resp := new(dns.Msg)
	resp.SetReply(request)
	resp.Authoritative = authoritative

	resp.Answer = records

	return resp
}

func ConvertToFqdn(name string) (string, error) {
	fqdn := dns.Fqdn(
		strings.TrimSpace(strings.ToLower(name)),
	)
	_, ok := dns.IsDomainName(fqdn)
	if !ok {
		return "", fmt.Errorf("invalid domain name: %q", name)
	}
	return fqdn, nil
}

func NewARecord(qname string, ttl uint32, ipv4 net.IP) (dns.RR, error) {
	fqdn, err := ConvertToFqdn(qname)
	if err != nil {
		return nil, fmt.Errorf("failed to convert hostname %q to FQDN: %w", qname, err)
	}
	//if len(ipv4) == net.IPv4len {
	//	return nil, fmt.Errorf("invalid IPv4 address: %q", ipv4)
	//}
	return &dns.A{
		Hdr: dns.RR_Header{
			Name:   fqdn,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    ttl,
		},
		A: ipv4,
	}, nil
}

func NewAAAARecord(qname string, ttl uint32, ipv6 net.IP) (dns.RR, error) {
	fqdn, err := ConvertToFqdn(qname)
	if err != nil {
		return nil, fmt.Errorf("failed to convert hostname %q to FQDN: %w", qname, err)
	}
	//if len(ipv6) == net.IPv6len {
	//	return nil, fmt.Errorf("invalid IPv6 address: %q", ipv6)
	//}
	return &dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   fqdn,
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    ttl,
		},
		AAAA: ipv6,
	}, nil
}

func NewCNAMERecord(qname string, ttl uint32, alias string) (dns.RR, error) {
	fqdn, err := ConvertToFqdn(qname)
	if err != nil {
		return nil, fmt.Errorf("failed to convert hostname %q to FQDN: %w", qname, err)
	}
	aliasFQDN, err := ConvertToFqdn(alias)
	if err != nil {
		return nil, fmt.Errorf("failed to convert alias %q to FQDN: %w", alias, err)
	}
	return &dns.CNAME{
		Hdr: dns.RR_Header{
			Name:   fqdn,
			Rrtype: dns.TypeCNAME,
			Class:  dns.ClassINET,
			Ttl:    ttl,
		},
		Target: aliasFQDN,
	}, nil
}
