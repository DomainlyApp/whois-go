/*
 * Copyright 2014-2020 Li Kexian
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Go module for domain and ip whois info query
 * https://www.likexian.com/
 */

package whois

import (
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"time"
)

const (
	// IANA_WHOIS_SERVER is iana whois server
	IANA_WHOIS_SERVER = "whois.iana.org"
	// DEFAULT_WHOIS_PORT is default whois port
	DEFAULT_WHOIS_PORT = "43"
)

// Version returns package version
func Version() string {
	return "1.4.0"
}

// Author returns package author
func Author() string {
	return "[Li Kexian](https://www.likexian.com/)"
}

// License returns package license
func License() string {
	return "Licensed under the Apache License 2.0"
}

// Whois do the whois query and returns whois info
func Whois(domain string, servers ...string) (result string, err error) {
	domain = strings.Trim(strings.TrimSpace(domain), ".")
	if domain == "" {
		err = fmt.Errorf("Domain is empty")
		return
	}

	if net.ParseIP(domain) == nil && !strings.Contains(domain, ".") {
		return query(domain, IANA_WHOIS_SERVER)
	}

	result, err = query(domain, servers...)
	if err != nil {
		return
	}

	server := getServer(result)
	if server == "" {
		return
	}

	data, err := query(domain, server)
	if err != nil {
		return
	}

	result += data

	return
}

// query do the query
func query(domain string, servers ...string) (result string, err error) {
	var server string
	if len(servers) == 0 || servers[0] == "" {
		ext := getExtension(domain)
		result, err = query(ext, IANA_WHOIS_SERVER)
		if err != nil {
			return "", fmt.Errorf("whois: query for whois server failed: %v", err)
		}
		server = getServer(result)
		if server == "" {
			return "", fmt.Errorf("whois: no whois server found")
		}
	} else {
		server = strings.ToLower(servers[0])
	}

	if server == "whois.arin.net" {
		domain = "n + " + domain
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(server, DEFAULT_WHOIS_PORT), time.Second*30)
	if err != nil {
		return "", fmt.Errorf("whois: connect to whois server failed: %v", err)
	}

	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(time.Second * 30))
	_, err = conn.Write([]byte(domain + "\r\n"))
	if err != nil {
		return "", fmt.Errorf("whois: send to whois server failed: %v", err)
	}

	buffer, err := ioutil.ReadAll(conn)
	if err != nil {
		return "", fmt.Errorf("whois: read from whois server failed: %v", err)
	}

	result = string(buffer)

	return
}

// getExtension returns extension of domain
func getExtension(domain string) string {
	ext := domain

	domains := strings.Split(domain, ".")
	if len(domains) > 1 {
		ext = domains[len(domains)-1]
	}

	if strings.Contains(ext, "/") {
		ext = strings.Split(ext, "/")[0]
	}

	return ext
}

// getServer returns server from whois data
func getServer(data string) string {
	tokens := []string{
		"Registrar WHOIS Server: ",
		"whois: ",
	}

	for _, token := range tokens {
		start := strings.Index(data, token)
		if start != -1 {
			start += len(token)
			end := strings.Index(data[start:], "\n")
			return strings.TrimSpace(data[start : start+end])
		}
	}

	return ""
}
