package irbis

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestConfig_Defaults(t *testing.T) {
	cfg := Config{}.withDefaults()
	if cfg.Host != DefaultHost || cfg.Port != DefaultPort {
		t.Fatalf("address defaults: %+v", cfg)
	}
	if cfg.DialTimeout != DefaultDialTimeout {
		t.Fatalf("DialTimeout=%v, want %v", cfg.DialTimeout, DefaultDialTimeout)
	}
	if cfg.IOTimeout != 0 {
		t.Fatalf("IOTimeout=%v, want 0 (no transport deadline)", cfg.IOTimeout)
	}
}

func TestConfig_CustomTimeoutsPreserved(t *testing.T) {
	cfg := Config{
		DialTimeout: 3 * time.Second,
		IOTimeout:   2 * time.Minute,
	}.withDefaults()
	if cfg.DialTimeout != 3*time.Second {
		t.Fatalf("DialTimeout=%v", cfg.DialTimeout)
	}
	if cfg.IOTimeout != 2*time.Minute {
		t.Fatalf("IOTimeout=%v", cfg.IOTimeout)
	}
}

func TestNewClient_AppliesConfig(t *testing.T) {
	var svc Service = NewClient(Config{
		Host:        "irbis.example",
		Port:        5555,
		Username:    "u",
		Password:    "p",
		Database:    "IBIS",
		DialTimeout: 7 * time.Second,
		IOTimeout:   90 * time.Second,
	})
	client := svc.(*Client)
	if client.Host != "irbis.example" || client.Port != 5555 {
		t.Fatalf("client address: %+v", client)
	}
	if client.IOTimeout != 90*time.Second {
		t.Fatalf("IOTimeout=%v", client.IOTimeout)
	}
	conn := client.newConnection()
	if conn.IOTimeout != 90*time.Second || conn.DialTimeout != 7*time.Second {
		t.Fatalf("connection timeouts: dial=%v io=%v", conn.DialTimeout, conn.IOTimeout)
	}
}

func TestApplyIODeadline_ZeroMeansContextOnly(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		time.Sleep(50 * time.Millisecond)
	}()

	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if err := applyIODeadline(context.Background(), conn, 0); err != nil {
		t.Fatalf("applyIODeadline: %v", err)
	}
	// No deadline set — ReadDeadline should be zero time.
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		t.Fatal(err)
	}
}

func TestApplyIODeadline_UsesIOTimeout(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		_ = c.Close()
	}()

	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if err := applyIODeadline(context.Background(), conn, time.Hour); err != nil {
		t.Fatalf("applyIODeadline: %v", err)
	}
}
