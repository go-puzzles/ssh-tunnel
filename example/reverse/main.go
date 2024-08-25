package main

import (
	"context"
	
	"github.com/go-puzzles/puzzles/pflags"
	"github.com/go-puzzles/ssh-tunnel/pkg/sshtunnel"
)

func main() {
	pflags.Parse()
	tunnel := sshtunnel.NewTunnel(&sshtunnel.SshConfig{
		User:         "hoven",
		HostName:     "10.11.43.115",
		IdentityFile: "/Users/yong/.ssh/id_rsa_cnns",
	})
	
	defer tunnel.Close()
	
	if err := tunnel.Reverse(context.TODO(), ":28081", "localhost:8080"); err != nil {
		panic(err)
	}
	
	tunnel.Wait()
}
