package cmd

import (
	"bytes"
	"errors"
	"sort"
	
	"github.com/go-puzzles/pflags"
	"github.com/go-puzzles/plog"
	"github.com/go-puzzles/ssh-tunnel/pkg/sshtunnel"
	"github.com/olekukonko/tablewriter"
)

var (
	localAddr  string
	remoteAddr string
)

var (
	port     = pflags.Int("port", 0, "Port for serivce")
	env      = pflags.String("env", "", "Enviroment name for looking up connection profile")
	profiles = pflags.Struct("profiles", []*ConnectionProfile{}, "Connection profiles")
)

type ConnectionProfile struct {
	EnvName string
	Host    *sshtunnel.SshConfig
}

func (cp *ConnectionProfile) PopulateDefault(identityFile string) {
	if cp.Host.IdentityFile == "" {
		cp.Host.IdentityFile = identityFile
	}
	if cp.Host.User == "" {
		cp.Host.User = "root"
	}
}

func (cp *ConnectionProfile) Validate() error {
	if cp.Host.HostName == "" {
		return errors.New("SSH hostName requred")
	}
	return nil
}

func newTunnel() *sshtunnel.SshTunnel {
	var profile *ConnectionProfile
	var allProfiles []*ConnectionProfile
	plog.PanicError(profiles(&allProfiles))
	
	for _, p := range allProfiles {
		if p.EnvName == env() {
			profile = p
			break
		}
	}
	
	if profile == nil {
		plog.Fatalf("no connection profile found. env=%v", env())
	}
	
	return sshtunnel.NewTunnel(profile.Host)
}

// prettyMaps formats a map to table format string.
func prettyMaps(m map[string][]string) string {
	buffer := &bytes.Buffer{}
	table := tablewriter.NewWriter(buffer)
	table.SetColWidth(400)
	
	type Record struct {
		Uuid   string
		Typ    string
		Tunnel string
	}
	var rs []*Record
	for name, us := range m {
		r := &Record{
			Uuid:   name,
			Typ:    us[0],
			Tunnel: us[1],
		}
		
		rs = append(rs, r)
	}
	sort.Slice(rs, func(i, j int) bool {
		return rs[i].Uuid < rs[j].Uuid
	})
	table.Append([]string{"UUID", "Tunnel-Type", "Tunnel"})
	for _, r := range rs {
		table.Append([]string{r.Uuid, r.Typ, r.Tunnel})
	}
	table.Render()
	return buffer.String()
}
