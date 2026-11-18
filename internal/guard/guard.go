package guard

import (
	"context"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"intra/internal/audit"
	"intra/internal/policy"
)

type Validator interface {
	Validate(actor, device, service string, action policy.Action, token string) bool
}

type Guard struct {
	Policy         *policy.Policy
	Validator      Validator // broker, consulted for grant/template access
	Audit          *audit.Client
	AllowHeaderID  bool // unit tests and lab only; production uses client certs
	LabViewerToken string
	ManifestDir    string
	ManifestKey    []byte
	proxies        map[string]*httputil.ReverseProxy
}

func New(p *policy.Policy, v Validator, au *audit.Client) (*Guard, error) {
	g := &Guard{Policy: p, Validator: v, Audit: au, proxies: map[string]*httputil.ReverseProxy{}}
	for name, svc := range p.Services {
		if svc.Kind == "postgres" || svc.Upstream == "" {
			continue
		}
		u, err := url.Parse(svc.Upstream)
		if err != nil {
			return nil, fmt.Errorf("service %s upstream: %w", name, err)
		}
		g.proxies[name] = httputil.NewSingleHostReverseProxy(u)
	}
	return g, nil
}

func identity(r *http.Request, allowHeader bool) (actor, device string) {
	if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
		cn := r.TLS.PeerCertificates[0].Subject.CommonName
		// CN form: actor/device
		if i := strings.Index(cn, "/"); i > 0 {
			return cn[:i], cn[i+1:]
		}
		return cn, ""
	}
	if allowHeader {
		return r.Header.Get("X-Intra-Actor"), r.Header.Get("X-Intra-Device")
	}
	return "", ""
}

func actionFor(r *http.Request, svc policy.Service) policy.Action {
	rest := r.URL.Path
	if i := strings.Index(strings.TrimPrefix(rest, "/svc/"), "/"); i >= 0 {
		rest = strings.TrimPrefix(rest, "/svc/")[i:]
	}
	best := -1
	bestAction := policy.ActionAdmin
	bestPrefix := ""
	for _, route := range svc.Routes {
		if route.Method != "" && route.Method != r.Method {
			continue
		}
		matched := rest == route.Prefix || (strings.HasPrefix(rest, route.Prefix) && (strings.HasSuffix(route.Prefix, "/") || rest[len(route.Prefix)] == '/'))
		if matched && len(route.Prefix) > best {
			best = len(route.Prefix)
			bestAction = route.Action
			bestPrefix = route.Prefix
		}
	}
	if best >= 0 {
		if (strings.Contains(rest, "/admin/") || strings.HasSuffix(rest, "/admin")) && !strings.Contains(bestPrefix, "/admin") {
			return policy.ActionAdmin
		}
		return bestAction
	}
	if strings.Contains(rest, "/admin") {
		return policy.ActionAdmin
	}
	if strings.HasSuffix(rest, "/export") {
		return policy.ActionExport
	}
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		return policy.ActionRead
	}
	return policy.ActionWrite
}

func (g *Guard) log(ctx context.Context, e audit.Event) error {
	if g.Audit != nil {
		return g.Audit.Append(ctx, e)
	}
	return nil
}

func (g *Guard) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(200)
			return
		}
		parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/svc/"), "/", 2)
		if !strings.HasPrefix(r.URL.Path, "/svc/") || len(parts) == 0 || parts[0] == "" {
			http.Error(w, "not found", 404)
			return
		}
		name := parts[0]
		svc, ok := g.Policy.Services[name]
		proxy, pok := g.proxies[name]
		if !ok || !pok {
			http.Error(w, "unknown service", 404)
			return
		}
		actor, device := identity(r, g.AllowHeaderID)
		act := actionFor(r, svc)
		d := g.Policy.Decide(actor, name, act)
		deny := func(reason string) {
			_ = g.log(r.Context(), audit.Event{Actor: actor, Device: device, Service: name, Action: string(act), Rule: d.Access, Result: "deny:" + reason})
			http.Error(w, "denied: "+reason, 403)
		}
		if actor == "" {
			deny("no identity")
			return
		}
		if g.ManifestDir != "" && g.Policy.Actors[actor].Kind == "robot" {
			var cert *x509.Certificate
			if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
				cert = r.TLS.PeerCertificates[0]
			}
			if err := VerifyRunManifest(g.ManifestDir, g.ManifestKey, actor, device, cert, time.Now()); err != nil {
				deny("run manifest: " + err.Error())
				return
			}
		}
		if !d.Allowed {
			deny(d.Reason)
			return
		}
		if d.Access == "grant" || d.Access == "template" {
			token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if g.Validator == nil || !g.Validator.Validate(actor, device, name, act, token) {
				deny("no active grant")
				return
			}
		}
		if err := g.log(r.Context(), audit.Event{Actor: actor, Device: device, Service: name, Action: string(act), Rule: d.Access, Result: "allow"}); err != nil {
			http.Error(w, "audit unavailable", 503)
			return
		}
		rest := "/"
		if len(parts) == 2 {
			rest += parts[1]
		}
		r2 := r.Clone(r.Context())
		r2.Header.Del("Authorization")
		r2.Header.Del("X-Intra-Actor")
		r2.Header.Del("X-Intra-Device")
		r2.Header.Del("X-Intra-Admin")
		if g.LabViewerToken != "" && (svc.Kind == "logs" || svc.Kind == "metrics") {
			r2.Header.Set("Authorization", "Bearer "+g.LabViewerToken)
		}
		r2.URL.Path = rest
		proxy.ServeHTTP(w, r2)
	})
}
