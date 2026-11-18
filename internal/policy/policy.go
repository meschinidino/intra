package policy

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

type Action string

const (
	ActionRead    Action = "read"
	ActionWrite   Action = "write"
	ActionAdmin   Action = "admin"
	ActionExport  Action = "export"
	ActionConnect Action = "connect"
	ActionEgress  Action = "egress"
)

type Actor struct {
	Name  string `json:"name"`
	Kind  string `json:"kind"` // human, robot, service, drill
	Owner string `json:"owner"`
}

type Route struct {
	Method string `json:"method"`
	Prefix string `json:"prefix"`
	Action Action `json:"action"`
}

type Service struct {
	Name      string   `json:"name"`
	Kind      string   `json:"kind"` // http, postgres, logs, metrics, saas, egress
	Sensitive bool     `json:"sensitive"`
	Upstream  string   `json:"upstream"`
	Actions   []Action `json:"actions"`
	Owner     string   `json:"owner"`
	Objects   []string `json:"objects,omitempty"` // database objects a read grant may reach
	Routes    []Route  `json:"routes,omitempty"`
}

type Rule struct {
	Actor         string   `json:"actor"`
	Service       string   `json:"service"`
	Actions       []Action `json:"actions"`
	Access        string   `json:"access"` // static, grant, template
	TemplateID    string   `json:"template_id,omitempty"`
	MaxTTLSeconds int      `json:"max_ttl_seconds,omitempty"`
}

type Template struct {
	ID            string   `json:"id"`
	Actor         string   `json:"actor"`
	Services      []string `json:"services"`
	Actions       []Action `json:"actions"`
	MaxTTLSeconds int      `json:"max_ttl_seconds"`
}

type Policy struct {
	Version   int                `json:"version"`
	Actors    map[string]Actor   `json:"actors"`
	Services  map[string]Service `json:"services"`
	Rules     []Rule             `json:"rules"`
	Templates []Template         `json:"templates,omitempty"`
	Egress    []EgressRule       `json:"egress,omitempty"`
}

type EgressRule struct {
	Actor string   `json:"actor"`
	Hosts []string `json:"hosts"`
	Ports []int    `json:"ports"`
}

type Decision struct {
	Allowed bool
	Access  string
	Rule    *Rule
	Reason  string
}

func Load(path string) (*Policy, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p Policy
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return &p, nil
}

func (p *Policy) Validate() error {
	if p.Version < 1 {
		return fmt.Errorf("policy version must be >= 1")
	}
	for name, s := range p.Services {
		if s.Name != name {
			return fmt.Errorf("service key %q does not match name %q", name, s.Name)
		}
		if s.Owner == "" {
			return fmt.Errorf("service %s has no owner", name)
		}
	}
	for i, r := range p.Rules {
		if _, ok := p.Actors[r.Actor]; !ok {
			return fmt.Errorf("rule %d: unknown actor %s", i, r.Actor)
		}
		svc, ok := p.Services[r.Service]
		if !ok {
			return fmt.Errorf("rule %d: unknown service %s", i, r.Service)
		}
		for _, a := range r.Actions {
			if !contains(svc.Actions, a) {
				return fmt.Errorf("rule %d: action %s not offered by %s", i, a, r.Service)
			}
		}
		switch r.Access {
		case "static", "grant", "template":
		default:
			return fmt.Errorf("rule %d: bad access %q", i, r.Access)
		}
		if r.Access == "template" && r.TemplateID == "" {
			return fmt.Errorf("rule %d: template access needs template_id", i)
		}
	}
	for _, t := range p.Templates {
		if _, ok := p.Actors[t.Actor]; !ok {
			return fmt.Errorf("template %s: unknown actor %s", t.ID, t.Actor)
		}
		if t.MaxTTLSeconds < 1 || t.MaxTTLSeconds > 7200 {
			return fmt.Errorf("template %s: ttl out of range", t.ID)
		}
		for _, name := range t.Services {
			svc, ok := p.Services[name]
			if !ok {
				return fmt.Errorf("template %s: unknown service %s", t.ID, name)
			}
			for _, a := range t.Actions {
				if !contains(svc.Actions, a) {
					return fmt.Errorf("template %s: action %s not offered by %s", t.ID, a, name)
				}
				if a != ActionRead && a != ActionEgress {
					return fmt.Errorf("template %s: only read and egress templates are supported", t.ID)
				}
			}
		}
	}
	for _, e := range p.Egress {
		if _, ok := p.Actors[e.Actor]; !ok {
			return fmt.Errorf("egress: unknown actor %s", e.Actor)
		}
		if len(e.Hosts) == 0 {
			return fmt.Errorf("egress for %s: no hosts", e.Actor)
		}
		for _, port := range e.Ports {
			if port < 1 || port > 65535 {
				return fmt.Errorf("egress for %s: bad port %d", e.Actor, port)
			}
		}
	}
	for _, t := range p.Templates {
		found := false
		for _, r := range p.Rules {
			if r.TemplateID == t.ID {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("template %s is not referenced by a rule", t.ID)
		}
	}
	return nil
}

func contains[T comparable](xs []T, x T) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

func (p *Policy) Decide(actor, service string, action Action) Decision {
	if _, ok := p.Actors[actor]; !ok {
		return Decision{Reason: "unknown actor"}
	}
	if _, ok := p.Services[service]; !ok {
		return Decision{Reason: "unknown service"}
	}
	for i := range p.Rules {
		r := &p.Rules[i]
		if r.Actor == actor && r.Service == service && contains(r.Actions, action) {
			return Decision{Allowed: true, Access: r.Access, Rule: r, Reason: "matched rule"}
		}
	}
	return Decision{Reason: "no matching rule"}
}

func (p *Policy) Template(id string) *Template {
	for i := range p.Templates {
		if p.Templates[i].ID == id {
			return &p.Templates[i]
		}
	}
	return nil
}

func (p *Policy) EgressAllowed(actor, host string, port int) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	for _, r := range p.Egress {
		if r.Actor != actor {
			continue
		}
		if len(r.Ports) > 0 && !contains(r.Ports, port) {
			continue
		}
		for _, h := range r.Hosts {
			h = strings.ToLower(h)
			if h == host || (strings.HasPrefix(h, "*.") && strings.HasSuffix(host, h[1:])) {
				return true
			}
		}
	}
	return false
}

// Compiled is the generated per-layer view. It is deterministic.
type Compiled struct {
	Version  int               `json:"version"`
	Hash     string            `json:"hash"`
	Mesh     []string          `json:"mesh_routes"`
	Guard    map[string]string `json:"guard_upstreams"`
	Broker   []Rule            `json:"broker_rules"`
	Services []Service         `json:"services"`
}

func (p *Policy) Compile() (*Compiled, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(p.Services))
	for n := range p.Services {
		names = append(names, n)
	}
	sort.Strings(names)
	c := &Compiled{Version: p.Version, Guard: map[string]string{}}
	for _, n := range names {
		s := p.Services[n]
		c.Services = append(c.Services, s)
		if s.Upstream != "" {
			c.Guard[s.Name] = s.Upstream
		}
		c.Mesh = append(c.Mesh, "actors->guard:"+s.Name)
	}
	c.Broker = append(c.Broker, p.Rules...)
	sort.Slice(c.Broker, func(i, j int) bool {
		if c.Broker[i].Actor != c.Broker[j].Actor {
			return c.Broker[i].Actor < c.Broker[j].Actor
		}
		return c.Broker[i].Service < c.Broker[j].Service
	})
	b, _ := json.Marshal(p)
	h := sha256.Sum256(b)
	c.Hash = hex.EncodeToString(h[:])
	return c, nil
}

var literalRE = regexp.MustCompile(`'(?:[^']|'')*'|\b\d+(?:\.\d+)?\b`)

// ClassifySQL returns an action and a literal-free fingerprint.
func ClassifySQL(sql string) (Action, string) {
	flat := strings.Join(strings.Fields(sql), " ")
	upper := strings.ToUpper(flat)
	action := ActionWrite
	words := map[string]bool{}
	for _, w := range strings.FieldsFunc(upper, func(r rune) bool {
		return !(r >= 'A' && r <= 'Z' || r == '_')
	}) {
		words[w] = true
	}
	switch {
	case words["SET_CONFIG"] || words["GRANT"] || words["REVOKE"] || words["ALTER"] || strings.HasPrefix(upper, "CREATE ROLE") || strings.HasPrefix(upper, "DROP ROLE"):
		action = ActionAdmin
	case words["INSERT"] || words["UPDATE"] || words["DELETE"] || words["MERGE"] || words["TRUNCATE"]:
		action = ActionWrite
	case strings.HasPrefix(upper, "COPY") && strings.Contains(upper, "TO STDOUT"):
		action = ActionExport
	case strings.HasPrefix(upper, "SELECT"), strings.HasPrefix(upper, "WITH"), strings.HasPrefix(upper, "SHOW"), strings.HasPrefix(upper, "EXPLAIN"):
		action = ActionRead
	}
	norm := literalRE.ReplaceAllString(flat, "?")
	h := sha256.Sum256([]byte(strings.ToLower(norm)))
	return action, hex.EncodeToString(h[:])[:16]
}
