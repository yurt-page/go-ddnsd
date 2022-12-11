package main

import (
	"bufio"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var listenAddrDns string
var listenApiAddr string
var selfDomain string
var selfIpStr string
var selfIp net.IP
var dnsRecords string
var verbose bool

func main() {
	flag.StringVar(&listenAddrDns, "listen-dns", ":53", "DNS listen IPv4 Address")
	flag.StringVar(&listenApiAddr, "listen-api", ":80", "API listen IPv4 Address")
	flag.StringVar(&selfDomain, "self-domain", "", "Domain of the DNS server e.g. dyn.jkl.mn.")
	flag.StringVar(&selfIpStr, "self-ip", "", "Public IP of the DNS server")
	flag.StringVar(&dnsRecords, "dns-records", "./dns.csv", "Saved DNS records")
	flag.BoolVar(&verbose, "verbose", false, "Verbose logging to stderr")
	flag.Parse()
	selfIp = net.ParseIP(selfIpStr)
	log.Printf("listen-dns: %s\n listen-api: %s\n self-domain: %s\n self-ip: %s\n verbose: %t\n dns-records: %s\n",
		listenAddrDns, listenApiAddr, selfDomain, selfIp.String(), verbose, dnsRecords)
	if selfDomain == "" || len(selfIpStr) == 0 {
		log.Fatal("Specify self-domain and self-ip")
	}

	if verbose {
		log.SetOutput(os.Stderr)
	}

	loadDnsRecords()
	go startDnsServer()
	go startApiServer()

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	for {
		select {
		case s := <-sig:
			if s == syscall.SIGHUP {
				loadDnsRecords()
			} else {
				os.Exit(0)
			}
		}
	}
}

func loadDnsRecords() {
	if dnsRecords == "" {
		return
	}
	dnsRecordsFile, err := os.Open(dnsRecords)
	if err != nil {
		log.Printf("err : %s\n", err.Error())
		return
	}
	defer dnsRecordsFile.Close()
	// clean old mappings
	mapRecordsLock.Lock()
	mapRecords = make(map[string]Records)
	mapRecordsLock.Unlock()
	scanner := bufio.NewScanner(dnsRecordsFile)
	for scanner.Scan() {
		line := scanner.Text()
		commentPos := strings.IndexByte(line, '#')
		if commentPos >= 0 {
			line = line[:commentPos]
		}
		line = strings.Replace(line, ",", " ", -1)
		fields := strings.Fields(line)
		fieldsLen := len(fields)
		if fieldsLen >= 2 {
			if verbose {
				log.Printf("%s\n", line)
			}
			domain := fields[0]
			ips := make([]net.IP, 0, fieldsLen-1)
			for i := 1; i < fieldsLen; i++ {
				ip := net.ParseIP(fields[i])
				if ip != nil {
					ips = append(ips, ip)
				}
			}
			UpdateDnsRecord(domain, ips)
		}
	}
}
