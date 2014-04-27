// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
package etcdutil

import (
	"errors"
	"github.com/coreos/go-etcd/etcd"
	"strings"
)

// Client is a wrapper around the etcd client
type Client struct {
	client *etcd.Client
}

// NewEtcdClient returns an *etcd.Client with a connection to named machines.
// It returns an error if a connection to the cluster cannot be made.
func NewEtcdClient(machines []string, cert, key string, caCert string) (*Client, error) {
	var c *etcd.Client
	if cert != "" && key != "" {
		c, err := etcd.NewTLSClient(machines, cert, key, caCert)
		if err != nil {
			return &Client{c}, err
		}
	} else {
		c = etcd.NewClient(machines)
	}
	success := c.SetCluster(machines)
	if !success {
		return &Client{c}, errors.New("cannot connect to etcd cluster: " + strings.Join(machines, ","))
	}
	return &Client{c}, nil
}

// GetValues queries etcd for keys prefixed by prefix.
// Etcd paths (keys) are translated into names more suitable for use in
// templates. For example if prefix were set to '/production' and one of the
// keys were '/nginx/port'; the prefixed '/production/nginx/port' key would
// be queried for. If the value for the prefixed key where 80, the returned map
// would contain the entry vars["nginx_port"] = "80".
func (c *Client) GetValues(prefix string, keys []string) (map[string]interface{}, error) {
	vars := make(map[string]interface{})
	for _, key := range keys {
		resp, err := c.client.Get(key, false, true)
		if err != nil {
			return vars, err
		}
		err = nodeWalk(resp.Node, prefix, vars)
		if err != nil {
			return vars, err
		}
	}
	return vars, nil
}

// nodeWalk recursively descends nodes, updating vars.
func nodeWalk(node *etcd.Node, prefix string, vars map[string]interface{}) error {
	if node != nil {
		key := node.Key
		if !node.Dir {
			vars[key] = node.Value
		} else {
			vars[key] = node.Nodes
			for _, node := range node.Nodes {
				nodeWalk(&node, prefix, vars)
			}
		}
	}
	return nil
}
