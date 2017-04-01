package handler

import (
	log "github.com/Sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"strconv"
	"time"
)

type Server struct {
	Host string
	Port int
	User string
	Pass string
}

func (server Server) Connect() bool {
	sshConfig := &ssh.ClientConfig{
		User: server.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(server.Pass),
		},
		Timeout:         time.Second * 2,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	connStr := server.Host + ":" + strconv.Itoa(server.Port)
	_, err := ssh.Dial("tcp", connStr, sshConfig)
	if err != nil {
		return false
	}
	return true
}

type Host struct {
	Host string
	Port int
}

type Auth struct {
	User string
	Pass string
}

func (host Host) Try(auth Auth) bool {
	s := Server{
		Host: host.Host,
		Port: host.Port,
		User: auth.User,
		Pass: auth.Pass,
	}
	log.Debugf("Trying %s:%s@%s:%d", auth.User, auth.Pass,
		host.Host, host.Port)
	if s.Connect() {
		log.Infof("Connected %s:%s@%s:%d", auth.User, auth.Pass,
			host.Host, host.Port)
		return true
	}
	return false
}
