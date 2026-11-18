package onboard

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

type ImportedPath struct {
	Source string `json:"source"`
	Peer   string `json:"peer"`
	Target string `json:"target"`
}

// ImportWireGuard reads peers and AllowedIPs from a plain wg config.
func ImportWireGuard(path string) ([]ImportedPath, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []ImportedPath
	peer := ""
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "PublicKey") {
			peer = strings.TrimSpace(strings.TrimPrefix(line, "PublicKey"))
			peer = strings.TrimPrefix(peer, "=")
			peer = strings.TrimSpace(peer)
			if len(peer) > 12 {
				peer = peer[:12]
			}
		}
		if strings.HasPrefix(line, "AllowedIPs") {
			val := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "AllowedIPs"), "="))
			for _, t := range strings.Split(val, ",") {
				out = append(out, ImportedPath{Source: "wireguard", Peer: peer, Target: strings.TrimSpace(t)})
			}
		}
	}
	return out, sc.Err()
}

type tailscaleDoc struct {
	ACLs []struct {
		Src []string `json:"src"`
		Dst []string `json:"dst"`
	} `json:"acls"`
	Grants []struct {
		Src []string `json:"src"`
		Dst []string `json:"dst"`
	} `json:"grants"`
}

// ImportTailscale reads acls/grants src and dst pairs from a policy export.
func ImportTailscale(path string) ([]ImportedPath, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var d tailscaleDoc
	if err := json.Unmarshal(b, &d); err != nil {
		return nil, fmt.Errorf("tailscale import: %w", err)
	}
	var out []ImportedPath
	for _, g := range append(d.ACLs, d.Grants...) {
		for _, s := range g.Src {
			for _, t := range g.Dst {
				out = append(out, ImportedPath{Source: "tailscale", Peer: s, Target: t})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return fmt.Sprint(out[i]) < fmt.Sprint(out[j]) })
	return out, nil
}
