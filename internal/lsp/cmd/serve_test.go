// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestListenParsing(t *testing.T) {
	tests := []struct {
		input, wantNetwork, wantAddr string
	}{
		{"127.0.0.1:0", "tcp", "127.0.0.1:0"},
		{"unix;/tmp/sock", "unix", "/tmp/sock"},
		{"auto", "auto", ""},
		{"auto;foo", "auto", "foo"},
	}

	for _, test := range tests {
		gotNetwork, gotAddr := parseAddr(test.input)
		if gotNetwork != test.wantNetwork {
			t.Errorf("network = %q, want %q", gotNetwork, test.wantNetwork)
		}
		if gotAddr != test.wantAddr {
			t.Errorf("addr = %q, want %q", gotAddr, test.wantAddr)
		}
	}
}

// TestServeRejectsImplicitAllInterfaces checks that -listen refuses an
// address with no host (such as ":0"), which would implicitly bind all
// network interfaces, while still accepting an explicit host.
func TestServeRejectsImplicitAllInterfaces(t *testing.T) {
	const implicitErr = "implicitly binds all network interfaces"

	// Use an already-cancelled context so that any server that does start
	// listening returns promptly instead of blocking the test.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	for _, addr := range []string{":0", "tcp;:0"} {
		s := &Serve{Address: addr, IdleTimeout: time.Millisecond, app: &Application{}}
		err := s.Run(ctx)
		if err == nil || !strings.Contains(err.Error(), implicitErr) {
			t.Errorf("Serve{Address: %q}.Run() = %v, want error containing %q", addr, err, implicitErr)
		}
	}

	for _, addr := range []string{"localhost:0", "tcp;localhost:0"} {
		s := &Serve{Address: addr, IdleTimeout: time.Millisecond, app: &Application{}}
		if err := s.Run(ctx); err != nil && strings.Contains(err.Error(), implicitErr) {
			t.Errorf("Serve{Address: %q}.Run() = %v, want no implicit-bind error", addr, err)
		}
	}
}

// TestServeNoPortFlag checks that the -port flag, which bound all network
// interfaces, is no longer offered by the serve command.
func TestServeNoPortFlag(t *testing.T) {
	typ := reflect.TypeOf(Serve{})
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Tag.Get("flag") == "port" {
			t.Errorf("Serve.%s provides the -port flag, which binds all network interfaces", field.Name)
		}
	}
}
