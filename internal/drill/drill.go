package drill

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"intra/internal/simulator"
)

type Target struct {
	Name      string `json:"name"`
	URL       string `json:"url"` // direct service URL that must deny
	Sensitive bool   `json:"sensitive"`
	Owner     string `json:"owner"`
}

type Result struct {
	PathID string `json:"path_id"`
	Target string `json:"target"`
	Owner  string `json:"owner"`
	Passed bool   `json:"passed"`
	Detail string `json:"detail"`
}

type Report struct {
	Results []Result `json:"results"`
	Score   float64  `json:"score_percent"`
	Blocked bool     `json:"release_blocked"`
}

// Run probes direct URLs with a drill identity. A pass means denial or no route.
func Run(ctx context.Context, targets []Target, canary string) Report {
	client := &http.Client{Timeout: 3 * time.Second}
	var rep Report
	var bad []string
	for _, t := range targets {
		id := "drill|" + t.Name + "|read"
		r := Result{PathID: id, Target: t.Name, Owner: t.Owner}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.URL, nil)
		if err != nil {
			r.Passed = true
			r.Detail = "request could not be built: " + err.Error()
			rep.Results = append(rep.Results, r)
			continue
		}
		req.Header.Set("X-Intra-Actor", "drill")
		resp, err := client.Do(req)
		if err != nil {
			r.Passed = true
			r.Detail = "no route: " + trimErr(err)
			rep.Results = append(rep.Results, r)
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		_ = resp.Body.Close()
		leaked := canary != "" && strings.Contains(string(body), canary)
		r.Passed = resp.StatusCode == 401 || resp.StatusCode == 403 || resp.StatusCode == 404
		if leaked {
			r.Passed = false
		}
		r.Detail = fmt.Sprintf("status %d canary=%v", resp.StatusCode, leaked)
		if !r.Passed {
			bad = append(bad, id)
			if t.Sensitive {
				rep.Blocked = true
			}
		}
		rep.Results = append(rep.Results, r)
	}
	score, ok := simulator.StandingScore(len(targets), bad)
	rep.Score = score
	if !ok {
		rep.Blocked = true
	}
	return rep
}

func trimErr(err error) string {
	s := err.Error()
	if len(s) > 120 {
		return s[:120]
	}
	return s
}

// DirectDialOpen reports whether a raw TCP route exists at all.
func DirectDialOpen(addr string) bool {
	c, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		return false
	}
	_ = c.Close()
	return true
}
