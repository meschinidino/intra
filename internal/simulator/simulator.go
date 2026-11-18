package simulator

import (
	"fmt"
	"sort"
	"strings"

	"intra/internal/audit"
	"intra/internal/policy"
)

type Path struct {
	ID      string `json:"id"`
	Actor   string `json:"actor"`
	Service string `json:"service"`
	Action  string `json:"action"`
	Access  string `json:"access"`
	Owner   string `json:"owner"`
}

func ReachPaths(p *policy.Policy) []Path {
	var out []Path
	for _, r := range p.Rules {
		for _, a := range r.Actions {
			out = append(out, Path{
				ID:      fmt.Sprintf("%s|%s|%s", r.Actor, r.Service, a),
				Actor:   r.Actor,
				Service: r.Service,
				Action:  string(a),
				Access:  r.Access,
				Owner:   p.Services[r.Service].Owner,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

type ReplayResult struct {
	Total       int     `json:"total"`
	FalseBlocks int     `json:"false_blocks"`
	Rate        float64 `json:"rate_percent"`
	Clean       bool    `json:"clean"`
}

// Replay scores recorded allow decisions against a candidate policy.
// A false block is a formerly allowed decision the candidate now denies.
func Replay(events []audit.Event, next *policy.Policy) ReplayResult {
	var r ReplayResult
	for _, e := range events {
		if e.Result != "allow" {
			continue
		}
		r.Total++
		services := strings.Split(e.Service, ",")
		blocked := false
		for _, svc := range services {
			if !next.Decide(e.Actor, svc, policy.Action(e.Action)).Allowed {
				blocked = true
			}
		}
		if blocked {
			r.FalseBlocks++
		}
	}
	if r.Total > 0 {
		r.Rate = 100 * float64(r.FalseBlocks) / float64(r.Total)
	}
	r.Clean = r.Total >= 200 && r.Rate < 1
	return r
}

type Draft struct {
	Actor   string `json:"actor"`
	Service string `json:"service"`
	Action  string `json:"action"`
	Count   int    `json:"count"`
}

// DraftRules groups observed allows into candidate rules. Deterministic: no model.
func DraftRules(events []audit.Event) []Draft {
	counts := map[string]int{}
	keys := map[string]Draft{}
	for _, e := range events {
		if e.Result != "allow" {
			continue
		}
		for _, svc := range strings.Split(e.Service, ",") {
			k := e.Actor + "|" + svc + "|" + e.Action
			counts[k]++
			keys[k] = Draft{Actor: e.Actor, Service: svc, Action: e.Action}
		}
	}
	var out []Draft
	for k, v := range keys {
		v.Count = counts[k]
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return fmt.Sprint(out[i]) < fmt.Sprint(out[j])
	})
	return out
}

// StandingScore implements the Stage 9 formula on distinct path IDs.
func StandingScore(totalPaths int, badPathIDs []string) (float64, bool) {
	if totalPaths <= 0 {
		return 100, false // empty inventory blocks release
	}
	uniq := map[string]bool{}
	for _, id := range badPathIDs {
		uniq[id] = true
	}
	score := 100 * float64(len(uniq)) / float64(totalPaths)
	if score > 100 {
		score = 100
	}
	return score, score <= 5
}
