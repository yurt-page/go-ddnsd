package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

var credentials = make(map[string]string)
var respCodeNoHost = []byte("nohost\n")
var respCodeBadAuth = []byte("badauth\n")

func startApiServer() {
	router := http.NewServeMux()
	router.HandleFunc("/nic/update", HandleDynDnsUpdateReq)

	apiServer := &http.Server{
		Addr:     listenApiAddr,
		Handler:  router,
		ErrorLog: log.New(io.Discard, "", 0), // most invalid requests are just exploits
	}
	err := apiServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Printf("http: Server shutdown %s\n", err.Error())
	}
}

func HandleDynDnsUpdateReq(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" {
		w.WriteHeader(405)
		return
	}
	defer func() {
		err := recover()
		if err != nil {
			log.Printf("%s\n", err)
		}
	}()
	headers := w.Header()
	headers.Set(`Content-Type`, `text/plain`)
	query := r.URL.Query()

	hostname := query.Get("hostname")
	domains := strings.Split(hostname, ",")
	if len(domains) == 0 {
		w.WriteHeader(200)
		w.Write(respCodeNoHost)
	}

	respCodes := make(net.Buffers, 0, len(domains))

	password := ""
	// take a password from the query param
	password = query.Get("token")
	// if no password then take it from Authorization
	if password == "" {
		_, password, _ = r.BasicAuth()
	}
	if password == "" {
		if verbose {
			log.Printf("%s no password\n", domains)
		}
		headers.Set(`WWW-Authenticate`, `Basic realm="`+selfDomain+`"`)
		w.WriteHeader(401)
		badAuth(w, domains, respCodes)
		return
	}

	badAuthCount := 0
	for i, domain := range domains {
		domain = strings.ToLower(strings.TrimSpace(domain))
		cred := credentials[domain]
		if cred != "" {
			if password != cred {
				// bad auth
				domains[i] = "badauth"
				badAuthCount++
				if verbose {
					log.Printf("domain %s badauth secret: %s\n", domain, password)
				}
			}
		} else {
			if verbose {
				log.Printf("domain %s register secret: %s\n", domain, password)
			}
			credentials[domain] = password
		}
	}
	// if all was bad then skip
	if badAuthCount == len(domains) {
		w.WriteHeader(200)
		badAuth(w, domains, respCodes)
		return
	}

	myip := query.Get("myip")
	addrs := getAddrs(myip, r)

	if verbose {
		log.Printf("domain %s secret: %s addrs: %s\n", domains, password, addrs)
	}

	for _, domain := range domains {
		//TODO "notfqdn\n"
		if domain == "" {
			respCodes = append(respCodes, respCodeNoHost)
			continue
		}
		if domain == "badauth" {
			respCodes = append(respCodes, respCodeBadAuth)
			continue
		}
		var respCode []byte
		if UpdateDnsRecord(domain, addrs) {
			respCode = []byte(fmt.Sprintf("good %s\n", addrs))
		} else {
			respCode = []byte(fmt.Sprintf("nochg %s\n", addrs))
		}
		respCodes = append(respCodes, respCode)
	}

	w.WriteHeader(200)
	respCodes.WriteTo(w)
	return
}

func badAuth(w http.ResponseWriter, domains []string, respCodes net.Buffers) {
	for range domains {
		respCodes = append(respCodes, respCodeBadAuth)
	}
	respCodes.WriteTo(w)
	return
}

func getAddrs(myip string, r *http.Request) []net.IP {
	if myip != "" {
		ips := strings.Split(myip, ",")
		addrs := make([]net.IP, 0, len(ips))
		for _, ip := range ips {
			parsedIP := net.ParseIP(strings.TrimSpace(ip))
			if parsedIP != nil {
				addrs = append(addrs, parsedIP)
			}
		}
		return addrs
	}
	remoteIpStr, _, _ := net.SplitHostPort(r.RemoteAddr)
	remoteIp := net.ParseIP(remoteIpStr)
	return []net.IP{remoteIp}
}
