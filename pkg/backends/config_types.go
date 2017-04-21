/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package backends

import (
	"fmt"
	"net"
	"strings"
)

// SRVRecord is a SRV-Record string
// for example _etcd-client._tcp.example.com.
type SRVRecord string

// GetNodesFromSRV returns the nodes stored in the record.
func (r SRVRecord) GetNodesFromSRV(scheme string) ([]string, error) {
	var nodes []string
	_, addrs, err := net.LookupSRV("", "", string(r))
	if err != nil {
		return nodes, err
	}

	for _, srv := range addrs {
		port := fmt.Sprintf("%d", srv.Port)
		host := strings.TrimRight(srv.Target, ".")
		host = net.JoinHostPort(host, port)
		if scheme != "" {
			host = fmt.Sprintf("%s://%s", scheme, host)
		}

		nodes = append(nodes, host)
	}
	return nodes, nil
}
