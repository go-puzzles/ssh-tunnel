/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"strings"
	
	"github.com/go-puzzles/puzzles/cores"
	grpcpuzzle "github.com/go-puzzles/puzzles/cores/puzzles/grpc-puzzle"
	grpcuipuzzle "github.com/go-puzzles/puzzles/cores/puzzles/grpcui-puzzle"
	"github.com/go-puzzles/puzzles/pflags"
	"github.com/go-puzzles/puzzles/plog"
	"github.com/go-puzzles/ssh-tunnel/internal/server"
	"github.com/go-puzzles/ssh-tunnel/sshtunnelpb"
	"github.com/spf13/cobra"
	
	"google.golang.org/grpc"
)

// ForwardCmd represents the connect command
var ForwardCmd = &cobra.Command{
	Use:   "forward -f <address:port> [flags] -t <address:port> [flags]",
	Short: "Proxy a locally accessible address to a remote address port",
	RunE: func(cmd *cobra.Command, args []string) error {
		pflags.Parse()
		
		tunnel := newTunnel()
		s := server.NewSSHTunnelServer(tunnel)
		
		ms := cores.NewPuzzleCore(
			cores.WithService("Ssh-Proxy"),
			grpcuipuzzle.WithCoreGrpcUI(),
			grpcpuzzle.WithCoreGrpcPuzzle(func(srv *grpc.Server) {
				sshtunnelpb.RegisterSshTunnelServer(srv, s)
			}),
			cores.WithDaemonWorker(func(ctx context.Context) error {
				in := &sshtunnelpb.ConnectRequest{
					Local:  localAddr,
					Remote: remoteAddr,
				}
				
				if strings.HasPrefix(in.Local, ":") {
					in.Local = "0.0.0.0" + in.Local
				}
				
				if strings.HasPrefix(in.Remote, ":") {
					in.Remote = "0.0.0.0" + in.Remote
				}
				
				resp, err := s.Forward(ctx, in)
				if err != nil {
					return err
				}
				
				table := make(map[string][]string)
				table[resp.Uuid] = append(
					table[resp.Uuid],
					[]string{string(server.Forward), fmt.Sprintf("%v -> %v", localAddr, remoteAddr)}...,
				)
				plog.Infof(prettyMaps(table))
				
				<-ctx.Done()
				tunnel.Close()
				return nil
			}),
		)
		return cores.Start(ms, port())
	},
}

func init() {
	ForwardCmd.Flags().StringVarP(&localAddr, "local", "l", "", "Local accessible address port")
	ForwardCmd.Flags().StringVarP(&remoteAddr, "remote", "r", "", "Remote address port")
	ForwardCmd.MarkFlagsRequiredTogether("local", "remote")
}
