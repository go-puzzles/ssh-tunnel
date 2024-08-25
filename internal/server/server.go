package server

import (
	"context"
	"fmt"
	"net/netip"
	"strings"
	
	"github.com/go-puzzles/puzzles/plog"
	"github.com/go-puzzles/ssh-tunnel/pkg/sshtunnel"
	"github.com/go-puzzles/ssh-tunnel/sshtunnelpb"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var _ sshtunnelpb.SshTunnelServer = (*Server)(nil)

type ConnectType string

const (
	Forward ConnectType = "Forward"
	Reverse ConnectType = "Reverse"
)

type connect struct {
	cancel func()
	typ    ConnectType
	local  netip.AddrPort
	remote netip.AddrPort
}

func (c *connect) Close() {
	c.cancel()
}

type Server struct {
	sshtunnelpb.UnimplementedSshTunnelServer
	
	sshTunnel *sshtunnel.SshTunnel
	cache     map[string]*connect
}

func NewSSHTunnelServer(tunnel *sshtunnel.SshTunnel) *Server {
	return &Server{
		sshTunnel: tunnel,
		cache:     make(map[string]*connect),
	}
}

func (s *Server) parseAddress(addr string) netip.AddrPort {
	return netip.MustParseAddrPort(addr)
}

func (s *Server) Forward(_ context.Context, in *sshtunnelpb.ConnectRequest) (*sshtunnelpb.ForwardReply, error) {
	ctx, cancel := context.WithCancel(context.Background())
	c := &connect{
		cancel: cancel,
		typ:    Forward,
		local:  s.parseAddress(in.Local),
		remote: s.parseAddress(in.Remote),
	}
	
	if err := s.sshTunnel.Forward(ctx, c.local.String(), c.remote.String()); err != nil {
		plog.Errorc(ctx, "SSH Forward %v -> %v error: %v", c.local, c.remote, err)
		return nil, errors.Wrap(err, "Forward")
	}
	
	uid, err := uuid.NewUUID()
	if err != nil {
		return nil, errors.Wrap(err, "NewUUID")
	}
	
	s.cache[uid.String()] = c
	
	plog.Infoc(ctx, "SSH Forward Tunnel %v -> %v success, Uid: %v", c.local, c.remote, uid)
	
	return &sshtunnelpb.ForwardReply{
		Uuid: uid.String(),
	}, nil
}

func (s *Server) Reverse(_ context.Context, in *sshtunnelpb.ConnectRequest) (*sshtunnelpb.ReverseReply, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	if strings.HasPrefix(in.Local, ":") {
		in.Local = "0.0.0.0" + in.Local
	}
	
	c := &connect{
		cancel: cancel,
		typ:    Reverse,
		local:  s.parseAddress(in.Local),
		remote: s.parseAddress(in.Remote),
	}
	
	if err := s.sshTunnel.Reverse(ctx, c.remote.String(), c.local.String()); err != nil {
		plog.Errorc(ctx, "SSH Reverse %v -> %v error: %v", c.remote, c.local, err)
		return nil, errors.Wrap(err, "Forward")
	}
	
	uid, err := uuid.NewUUID()
	if err != nil {
		return nil, errors.Wrap(err, "NewUUID")
	}
	
	s.cache[uid.String()] = c
	
	plog.Infoc(ctx, "SSH Reverse Tunnel %v -> %v success, Uid: %v", c.remote, c.local, uid)
	
	return &sshtunnelpb.ReverseReply{
		Uuid: uid.String(),
	}, nil
}

func (s *Server) Disconnect(ctx context.Context, in *sshtunnelpb.DisconnectRequest) (*sshtunnelpb.DisconnectReply, error) {
	c, exists := s.cache[in.Uuid]
	if !exists {
		return nil, fmt.Errorf("uid: %v connect not exists", in.Uuid)
	}
	
	c.Close()
	
	plog.Infoc(ctx, "Tunnel[%v] close success!", in.Uuid)
	return &sshtunnelpb.DisconnectReply{}, nil
}

func (s *Server) ListConnect(ctx context.Context, in *sshtunnelpb.ListConnectRequest) (*sshtunnelpb.ListConnectReply, error) {
	out := &sshtunnelpb.ListConnectReply{
		Connects: make([]*sshtunnelpb.Connect, 0, len(s.cache)),
	}
	
	for key, value := range s.cache {
		out.Connects = append(out.Connects, &sshtunnelpb.Connect{
			Uuid:        key,
			ConnectType: string(value.typ),
			Local:       value.local.String(),
			Remote:      value.remote.String(),
		})
	}
	
	return out, nil
}
