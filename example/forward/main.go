package main

import (
	"context"
	
	"github.com/go-puzzles/ssh-tunnel/pkg/sshtunnel"
)

func main() {
	tunnel := sshtunnel.NewTunnel(&sshtunnel.SshConfig{
		User:         "hoven",
		HostName:     "10.11.43.115",
		IdentityFile: "/Users/yong/.ssh/id_rsa",
	})
	
	if err := tunnel.Forward(context.TODO(), "localhost:28080", "localhost:80"); err != nil {
		panic(err)
	}
	
	tunnel.Wait()
}
