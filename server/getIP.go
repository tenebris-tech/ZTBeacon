//
// Copyright (c) 2023 Tenebris Technologies Inc.
// See LICENSE for further information.
//

package server

import (
	"net/http"
	"strings"
)

// getIP gets a requests IP address by reading the forwarded-for
// header (for proxies or load balancers) and falls back to use the remote address.
func (c *Config) getIP(r *http.Request) string {
	var s = ""
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		s = forwarded
	} else {
		s = r.RemoteAddr
	}

	// Clean up, remove port number
	if len(s) > 0 {
		if strings.HasPrefix(s, "[") {
			// IPv6 address
			t := strings.Split(s, "]")
			s = t[0][1:]
		} else {
			// IPv4 - hack off port number
			t := strings.Split(s, ":")
			s = t[0]
		}
	}
	return s
}
