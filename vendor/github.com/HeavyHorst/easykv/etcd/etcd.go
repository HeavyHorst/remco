/*
 * This file is part of easyKV.
 * © 2016 The easyKV Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package etcd

import (
	"errors"

	"github.com/HeavyHorst/easykv"
	"github.com/HeavyHorst/easykv/etcd/etcdv2"
	"github.com/HeavyHorst/easykv/etcd/etcdv3"
)

// ErrUnknownAPILevel is returned if no valid api level is given
var ErrUnknownAPILevel = errors.New("unknown etcd api level - must be 2 or 3")

// New returns an *etcd{2,3}.Client with a connection to named machines.
func New(machines []string, opts ...Option) (easykv.ReadWatcher, error) {
	var options Options
	for _, o := range opts {
		o(&options)
	}
	options.Nodes = machines

	ba := false
	if options.Auth.Password != "" && options.Auth.Username != "" {
		ba = true
	}

	if options.Version == 3 {
		return etcdv3.NewEtcdClient(options.Nodes, options.TLS.ClientCert, options.TLS.ClientKey, options.TLS.ClientCaKeys, ba, options.Auth.Username, options.Auth.Password)
	}

	if options.Version == 2 {
		return etcdv2.NewEtcdClient(options.Nodes, options.TLS.ClientCert, options.TLS.ClientKey, options.TLS.ClientCaKeys, ba, options.Auth.Username, options.Auth.Password)
	}

	return nil, ErrUnknownAPILevel
}
