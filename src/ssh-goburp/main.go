package main

import (
	"bufio"
	"flag"
	"handler"
	"io"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
)

func main() {
	// target
	var targethost = flag.String("h", "", "Target Host")
	var targetport = flag.Int("p", 22, "Target Port")
	// login user
	var username = flag.String("u", "", "Login User / Username lists")
	// passwords
	var password = flag.String("P", "", "Login Pass / Password lists")
	// debug mode
	var debug = flag.Bool("d", false, "Debug Mode")

	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}
	log.Debugf("host:[%v], port:[%v], users:[%v], pass:[%v]", *targethost, *targetport, *username, *password)

	var (
		host      string
		port      int
		usernames []string
		passwords []string
	)

	if *username == "" || *password == "" || *targethost == "" || *targetport < 0 {
		log.Info("username")
		flag.Usage()
		os.Exit(1)
	}

	host = *targethost
	port = *targetport
	target := handler.Host{
		Host: host,
		Port: port,
	}
	log.Infof("Target is %s:%d", target.Host, target.Port)

	if !PathExist(*username) {
		usernames = append(usernames, *username)
	} else {
		usernames = readFile(*username)
	}

	if !PathExist(*password) {
		passwords = append(passwords, *password)
	} else {
		passwords = readFile(*password)
	}

	log.Infof("Load [%v] passwords", len(passwords))
	stime := time.Now()
	log.Infof("Starting at %v", stime)

	tries := make(chan int, 1)
	tries <- 1
	success := make(chan int, 1)
	success <- 0
	passwd := make(chan string, 1)
	total := len(usernames) * len(passwords)
	for u := 0; u < len(usernames); u++ {
		for p := 0; p < len(passwords); p++ {
			auth := handler.Auth{
				User: usernames[u],
				Pass: passwords[p],
			}
			if <-success == 1 {
				success <- 1
				break
			}
			success <- 0
			go func(auth handler.Auth, success chan int, tries chan int, passwd chan string) {
				if <-success == 1 {
					success <- 1
					return
				}
				success <- 0
				if target.Try(auth) {
					log.Info(time.Now())
					log.Infof("Auth [%v] Connected!", auth)
					<-success
					success <- 1
					passwd <- auth.Pass
					return
				}
				t := <-tries
				if t == total {
					log.Error("Password not found.")
					<-success
					success <- 1
					passwd <- "Password not found."
					return
				}
				tries <- t + 1
			}(auth, success, tries, passwd)
			time.Sleep(25000 * time.Microsecond)
		}
	}
	p := <-passwd
	if p != "Password not found." {
		log.Infof("Password Found: [%s]", p)
	}

	ftime := time.Now()
	log.Infof("Finished at %v", ftime)
	ttime := time.Since(stime)
	log.Infof("Used: [%v]", ttime.Seconds())

}

func PathExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func readFile(filename string) []string {
	appendTo := make([]string, 0)
	f, _ := os.OpenFile(filename, os.O_RDONLY, 0666)
	rb := bufio.NewReader(f)
	for {
		line, _, err := rb.ReadLine()
		if err == io.EOF {
			break
		}
		appendTo = append(appendTo, string(line))
	}
	f.Close()
	return appendTo
}
