// Copyright 2016 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package tikv

import (
	"time"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/juju/errors"
	"github.com/pingcap/kvproto/pkg/tikvpb"
	goctx "golang.org/x/net/context"
	"google.golang.org/grpc"
)

const defaultBufSize = 4 * 1024

// Conn is a simple wrapper of grpc.ClientConn.
type Conn struct {
	*grpc.ClientConn
	rawStream tikvpb.Tikv_RawStreamClient
	addr      string
	closed    bool
}

// NewConnection creates a Conn with dial timeout.
func NewConnection(addr string, dialTimeout time.Duration) (*Conn, error) {
	conn, err := grpc.Dial(
		addr,
		grpc.WithInsecure(),
		grpc.WithTimeout(dialTimeout),
		grpc.WithInitialWindowSize(grpcInitialWindowSize),
		grpc.WithInitialConnWindowSize(grpcInitialConnWindowSize),
		grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor))
	if err != nil {
		return nil, errors.Trace(err)
	}

	client := tikvpb.NewTikvClient(conn)
	rawStream, err := client.RawStream(goctx.TODO())
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &Conn{
		ClientConn: conn,
		rawStream:  rawStream,
		addr:       addr,
		closed:     false,
	}, nil
}

// Close closes the grpc.ClientConn.
func (c *Conn) Close() {
	if c.closed {
		return
	}
	c.closed = true
	if c.ClientConn != nil {
		c.ClientConn.Close()
		c.ClientConn = nil
	}
}
