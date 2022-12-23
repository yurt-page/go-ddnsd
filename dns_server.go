package main

import (
	"github.com/miekg/dns"
	"log"
	"net"
	"strings"
	"sync"
)

// 5 min
var defaultTtl uint32 = 600

type Records struct {
	Password string
	Rrs      []dns.RR
}

var mapRecords = make(map[string]Records)
var mapRecordsLock = sync.RWMutex{}

func startDnsServer() {
	go func() {
		dns.HandleFunc(selfDomain, handleDnsRequest)
		srv := &dns.Server{Addr: listenAddrDns, Net: "udp"}
		err := srv.ListenAndServe()
		if err != nil {
			log.Fatalf("DNS serv failed: %s\n", err.Error())
		}
	}()
	go func() {
		dns.HandleFunc(selfDomain, handleDnsRequest)
		srv := &dns.Server{Addr: listenAddrDns, Net: "tcp"}
		err := srv.ListenAndServe()
		if err != nil {
			log.Fatalf("DNS serv failed: %s\n", err.Error())
		}
	}()
}

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	//FIXME check questions
	domain := r.Question[0].Name
	fqdn := strings.ToLower(domain)
	if fqdn == selfDomain {
		handleSelfDnsRequest(w, r)
		return
	}
	if verbose {
		log.Printf("handleDnsRequest %s\n", domain)
	}

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true

	mapRecordsLock.RLock()
	records, found := mapRecords[fqdn]
	mapRecordsLock.RUnlock()
	if found {
		m.Answer = records.Rrs
	} else {
		m.Rcode = dns.RcodeNameError
	}

	err := w.WriteMsg(m)
	if err != nil {
		log.Print(err)
		return
	}

	if verbose {
		fromIp, _, _ := net.SplitHostPort(w.RemoteAddr().String())
		if found {
			log.Printf("%s %s: records %s\n", fromIp, domain, records.Rrs)
		} else {
			log.Printf("%s %s: not found\n", fromIp, domain)
		}
	}
}

func handleSelfDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	//TODO check questions
	domain := r.Question[0].Name
	if verbose {
		log.Printf("handleSelfDnsRequest %s\n", domain)
	}
	//TODO self-ip IPv6 AAAA
	rr1 := &dns.A{
		A:   selfIp,
		Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: defaultTtl},
	}
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.Answer = []dns.RR{rr1}
	err := w.WriteMsg(m)
	if err != nil {
		log.Print(err)
		return
	}
}

func UpdateDnsRecord(domain string, ips []net.IP) bool {
	if verbose {
		log.Printf("%s -> %s\n", domain, ips)
	}
	if domain[len(domain)-1] != '.' {
		domain += "."
	}
	rrs := make([]dns.RR, 0, len(ips))
	for _, ip := range ips {
		ip4 := ip.To4()
		isIp4 := ip4 != nil
		var rr dns.RR
		if isIp4 {
			rr = &dns.A{A: ip4, Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: defaultTtl}}
		} else {
			rr = &dns.AAAA{AAAA: ip, Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: defaultTtl}}
		}
		rrs = append(rrs, rr)
	}
	mapRecordsLock.Lock()
	records := Records{Password: "", Rrs: rrs}
	mapRecords[domain] = records
	mapRecordsLock.Unlock()
	return true
}
