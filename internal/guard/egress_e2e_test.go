package guard

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"intra/internal/audit"
	"intra/internal/policy"
)

type allowValidator struct{}

func (allowValidator) Validate(string, string, string, policy.Action, string) bool { return true }

func TestEgressConnectValidatedIPAndAudit(t *testing.T) {
	echo, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer echo.Close()
	go func() {
		c, err := echo.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		buf := make([]byte, 4)
		_, _ = c.Read(buf)
		_, _ = c.Write([]byte("pong"))
	}()

	dir := t.TempDir()
	w, err := audit.OpenWriter(filepath.Join(dir, "audit.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	asrv := httptest.NewServer((&audit.Server{W: w}).Handler())
	defer asrv.Close()

	p := &policy.Policy{
		Version: 1,
		Actors:  map[string]policy.Actor{"robot": {Name: "robot", Kind: "robot", Owner: "f"}},
		Egress:  []policy.EgressRule{{Actor: "robot", Hosts: []string{"example.com"}, Ports: []int{443}}},
	}
	var dialed string
	e := &Egress{
		Policy: p, Validator: allowValidator{}, Audit: &audit.Client{Base: asrv.URL}, AllowHeaderID: true,
		LookupIP: func(string) ([]net.IP, error) { return []net.IP{net.ParseIP("93.184.216.34")}, nil },
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dialed = addr
			var d net.Dialer
			return d.DialContext(ctx, network, echo.Addr().String())
		},
	}
	srv := httptest.NewServer(e.Handler())
	defer srv.Close()

	conn, err := net.Dial("tcp", srv.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_, _ = fmt.Fprintf(conn, "CONNECT example.com:443 HTTP/1.1\r\nHost: example.com:443\r\nX-Intra-Actor: robot\r\nAuthorization: Bearer grant-token\r\n\r\n")
	br := bufio.NewReader(conn)
	status, err := br.ReadString('\n')
	if err != nil || status == "" {
		t.Fatalf("connect status %q err %v", status, err)
	}
	// Drain header lines to the blank line.
	for {
		line, err := br.ReadString('\n')
		if err != nil || line == "\r\n" {
			break
		}
	}
	if _, err := conn.Write([]byte("ping")); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 4)
	if _, err := br.Read(buf); err != nil || string(buf) != "pong" {
		t.Fatalf("echo %q err %v", buf, err)
	}
	if dialed != "93.184.216.34:443" {
		t.Fatalf("dialed %q, want validated IP", dialed)
	}
	events, err := audit.ReadAll(filepath.Join(dir, "audit.jsonl"))
	if err != nil || len(events) != 1 || events[0].Result != "allow" {
		t.Fatalf("audit events %+v err %v", events, err)
	}
}
