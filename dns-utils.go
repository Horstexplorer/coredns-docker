package docker

import (
	"net"

	"github.com/miekg/dns"
)

func BuildDNSResponse(request *dns.Msg, authoritative bool, records []dns.RR) *dns.Msg {
	resp := new(dns.Msg)
	resp.SetReply(request)
	resp.Authoritative = authoritative

	resp.Answer = records

	return resp
}

func ARecord(qname string, ttl uint32, ipv4 string) dns.RR {
	return &dns.A{
		Hdr: dns.RR_Header{
			Name:   qname,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    ttl,
		},
		A: net.ParseIP(ipv4),
	}
}

func AAAARecord(qname string, ttl uint32, ipv6 string) dns.RR {
	return &dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   qname,
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    ttl,
		},
		AAAA: net.ParseIP(ipv6),
	}
}

func CNAMERecord(qname string, ttl uint32, alias string) dns.RR {
	return &dns.CNAME{
		Hdr: dns.RR_Header{
			Name:   qname,
			Rrtype: dns.TypeCNAME,
			Class:  dns.ClassINET,
			Ttl:    ttl,
		},
		Target: alias,
	}
}
