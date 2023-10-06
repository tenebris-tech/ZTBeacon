//
// Copyright (c) 2023 Tenebris Technologies Inc.
// See LICENSE for further information.
//

package server

// deny checks if an IP address is in the deny list
func (c *Config) deny(ip string) bool {
	c.denyList.mu.RLock()
	defer c.denyList.mu.RUnlock()

	for _, v := range c.denyList.ipList {
		if v == ip {
			return true
		}
	}

	return false
}

// AddDeny adds an IP address to the deny list
func (c *Config) AddDeny(ip string) {
	c.denyList.mu.Lock()
	defer c.denyList.mu.Unlock()

	c.denyList.ipList = append(c.denyList.ipList, ip)
}
