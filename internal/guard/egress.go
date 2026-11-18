package guard

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"intra/internal/audit"
	"intra/internal/policy"
)

// Egress is a deterministic forward proxy. For CONNECT it can see host and
// port only. It never intercepts TLS and never claims body inspection.
type Egress struct {
	Policy        *policy.Policy
	Dialer        *net.Dialer
	Validator     Validator
	Audit         *audit.Client
	AllowHeaderID bool // tests only
	LookupIP      func(host string) ([]net.IP, error)
	DialContext   func(ctx context.Context, network, addr string) (net.Conn, error)
}

var nonGlobalCIDRs = func() []*net.IPNet {
	var out []*net.IPNet
	for _, s := range []string{
		"0.0.0.0/8", "10.0.0.0/8", "100.64.0.0/10", "127.0.0.0/8", "169.254.0.0/16",
		"172.16.0.0/12", "192.0.0.0/24", "192.0.2.0/24", "192.88.99.0/24", "192.168.0.0/16",
		"198.18.0.0/15", "198.51.100.0/24", "203.0.113.0/24", "224.0.0.0/4", "240.0.0.0/4",
		"::1/128", "64:ff9b::/96", "64:ff9b:1::/48", "100::/64", "2001::/32", "2001:db8::/32", "fc00::/7", "fe80::/10", "ff00::/8",
	} {
		_, n, err := net.ParseCIDR(s)
		if err == nil {
			out = append(out, n)
		}
	}
	return out
}()

func privateIP(ip net.IP) bool {
	if ip == nil || ip.IsUnspecified() {
		return true
	}
	for _, n := range nonGlobalCIDRs {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

func (e *Egress) resolve(host string) (net.IP, error) {
	lookup := e.LookupIP
	if lookup == nil {
		lookup = net.LookupIP
	}
	if ip := net.ParseIP(host); ip != nil {
		if privateIP(ip) {
			return nil, fmt.Errorf("non-global address")
		}
		return ip, nil
	}
	ips, err := lookup(host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("no addresses")
	}
	for _, ip := range ips {
		if privateIP(ip) {
			return nil, fmt.Errorf("host resolves to a non-global address")
		}
	}
	return ips[0], nil
}

func (e *Egress) dial(ctx context.Context, ip net.IP, port int) (net.Conn, error) {
	addr := net.JoinHostPort(ip.String(), strconv.Itoa(port))
	if e.DialContext != nil {
		return e.DialContext(ctx, "tcp", addr)
	}
	d := e.Dialer
	if d == nil {
		d = &net.Dialer{Timeout: 10 * time.Second}
	}
	return d.DialContext(ctx, "tcp", addr)
}

func resolveGlobal(host string) (net.IP, error) {
	if ip := net.ParseIP(host); ip != nil {
		if privateIP(ip) {
			return nil, fmt.Errorf("non-global address")
		}
		return ip, nil
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("no addresses")
	}
	for _, ip := range ips {
		if privateIP(ip) {
			return nil, fmt.Errorf("host resolves to a non-global address")
		}
	}
	return ips[0], nil
}

func (e *Egress) allowed(actor, hostport string) (net.IP, int, bool) {
	host, portText, err := net.SplitHostPort(hostport)
	if err != nil {
		host = hostport
		portText = "443"
	}
	port, _ := strconv.Atoi(portText)
	if e.Policy == nil || !e.Policy.EgressAllowed(actor, host, port) {
		return nil, port, false
	}
	ip, err := e.resolve(host)
	if err != nil {
		return nil, port, false
	}
	return ip, port, true
}

func (e *Egress) record(ctx context.Context, actor, device, host, result string) error {
	if e.Audit == nil {
		return fmt.Errorf("audit unavailable")
	}
	return e.Audit.Append(ctx, audit.Event{Actor: actor, Device: device, Service: "internet", Action: "egress", Rule: host, Result: result})
}

func (e *Egress) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor, device := identity(r, e.AllowHeaderID)
		target := r.Host
		if r.Method != http.MethodConnect && r.URL != nil && r.URL.Host != "" {
			target = r.URL.Host
		}
		deny := func(code int, msg string) {
			_ = e.record(r.Context(), actor, device, target, "deny:"+msg)
			http.Error(w, msg, code)
		}
		if actor == "" {
			deny(403, "no identity")
			return
		}
		if e.Validator != nil {
			token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !e.Validator.Validate(actor, device, "internet", policy.ActionEgress, token) {
				deny(403, "no active egress grant")
				return
			}
		}
		dialer := e.Dialer
		if dialer == nil {
			dialer = &net.Dialer{Timeout: 10 * time.Second}
		}
		if r.Method == http.MethodConnect {
			ip, port, ok := e.allowed(actor, r.Host)
			if !ok {
				deny(403, "egress denied")
				return
			}
			if err := e.record(r.Context(), actor, device, r.Host, "allow"); err != nil {
				http.Error(w, "audit unavailable", 503)
				return
			}
			dst, err := e.dial(r.Context(), ip, port)
			if err != nil {
				http.Error(w, "dial failed", 502)
				return
			}
			hij, ok := w.(http.Hijacker)
			if !ok {
				_ = dst.Close()
				http.Error(w, "hijack unsupported", 500)
				return
			}
			client, buf, err := hij.Hijack()
			if err != nil {
				_ = dst.Close()
				return
			}
			_, _ = buf.WriteString("HTTP/1.1 200 Connection Established\r\n\r\n")
			_ = buf.Flush()
			go func() { _, _ = io.Copy(dst, client); _ = dst.Close() }()
			_, _ = io.Copy(client, dst)
			_ = client.Close()
			return
		}
		if r.URL == nil || r.URL.Host == "" {
			deny(400, "absolute URL required")
			return
		}
		if r.URL.Scheme == "https" {
			deny(400, "use CONNECT for https")
			return
		}
		hostport := r.URL.Host
		if !strings.Contains(hostport, ":") {
			hostport = net.JoinHostPort(hostport, "80")
		}
		if _, _, ok := e.allowed(actor, hostport); !ok {
			deny(403, "egress denied")
			return
		}
		if err := e.record(r.Context(), actor, device, hostport, "allow"); err != nil {
			http.Error(w, "audit unavailable", 503)
			return
		}
		transport := &http.Transport{
			Proxy: nil,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				host, portText, err := net.SplitHostPort(addr)
				if err != nil {
					return nil, err
				}
				ip, err := e.resolve(host)
				if err != nil {
					return nil, err
				}
				p, _ := strconv.Atoi(portText)
				return e.dial(ctx, ip, p)
			},
		}
		out := r.Clone(r.Context())
		out.RequestURI = ""
		out.Header.Del("Authorization")
		out.Header.Del("X-Intra-Actor")
		out.Header.Del("X-Intra-Device")
		resp, err := transport.RoundTrip(out)
		if err != nil {
			http.Error(w, fmt.Sprintf("upstream: %v", err), 502)
			return
		}
		defer resp.Body.Close()
		for k, vs := range resp.Header {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	})
}
