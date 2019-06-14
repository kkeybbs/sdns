package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"github.com/miekg/dns"
	"gopkg.in/yaml.v2"
)

type Server struct {
	Rule
	Addr    string
	Servers []string

	client dns.Client
	self   string
}

func (s *Server) createMsg(r *dns.Msg, rr dns.RR) *dns.Msg {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false
	if rr != nil {
		m.Answer = append(m.Answer, rr)
	}
	return m
}

func (s *Server) matchRules(name string) string {
	name = strings.TrimSuffix(name, ".")
	if ip := s.Match(name); ip != nil {
		return *ip
	}
	return ""
}

func (s *Server) answer(name string) dns.RR {
	if ip := s.matchRules(name); ip != "" {
		if ip == "self" {
			ip = s.self
		}
		rr, err := dns.NewRR(fmt.Sprintf("%s A %s", name, ip))
		if err == nil {
			return rr
		}
	}
	return nil
}

func (s *Server) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	tag := fmt.Sprintf("op=%d qt=%d name=%s", r.Opcode, r.Question[0].Qtype, r.Question[0].Name)
	switch r.Opcode {
	case dns.OpcodeQuery:
		for _, q := range r.Question {
			switch q.Qtype {
			case dns.TypeA:
				if rr := s.answer(q.Name); rr != nil {
					m := s.createMsg(r, rr)
					w.WriteMsg(m)
					return
				}
			}
		}
	}
	for i, addr := range s.Servers {
		m, rtt, err := s.client.Exchange(r, addr)
		if err == nil {
			fmt.Println("query ok:", addr, tag, rtt)
			for _, answer := range m.Answer {
				fmt.Println("\t", answer.String())
			}
			w.WriteMsg(m)
			if i > 0 {
				s.Servers = append(s.Servers[i:], s.Servers[0:i]...)
			}
			return
		}
		fmt.Println("query failed:", addr, tag, err)
	}
	m := s.createMsg(r, nil)
	w.WriteMsg(m)
}

func (s *Server) getSelfIP() (ip string) {
	targets := []string{"8.8.8.8", "10.10.10.10"}
	var conn net.Conn
	var err error
	for _, target := range targets {
		if conn, err = net.Dial("udp4", target+":80"); err == nil {
			ip = conn.LocalAddr().(*net.UDPAddr).IP.String()
			conn.Close()
			return
		}
	}
	return ""
}

func (s *Server) Load(path string) error {
	if data, err := ioutil.ReadFile(path); err != nil {
		return err
	} else if err = yaml.Unmarshal(data, s); err != nil {
		return err
	}
	if s.Addr == "" {
		s.Addr = "127.0.0.1:53"
	}
	if s.Servers == nil || len(s.Servers) == 0 {
		return errors.New("no forward server")
	}
	if s.self = s.getSelfIP(); s.self == "" {
		return errors.New("get self ip failed")
	}
	if err := s.Compile(); err != nil {
		return err
	}
	for i, value := range s.Servers {
		if !strings.Contains(value, ":") {
			s.Servers[i] = value + ":53"
		}
	}
	return nil
}

func (s *Server) Run() {
	done := make(chan error)
	go func() {
		done <- dns.ListenAndServe(s.Addr, "udp", s)
	}()
	go func() {
		done <- dns.ListenAndServe(s.Addr, "tcp", s)
	}()
	err := <-done
	if err != nil {
		fmt.Println("failed to serve", err)
	}
}

func main() {
	var config = "sdns.yaml"
	var test string
	flag.StringVar(&config, "c", config, "config path")
	flag.StringVar(&test, "t", "", "test match rule")
	flag.Parse()
	s := &Server{}
	if err := s.Load(config); err != nil {
		fmt.Println("load config failed:", err)
		os.Exit(1)
		return
	}
	if test == "" {
		s.Run()
		return
	}
	if ip := s.Match(test); ip != nil {
		fmt.Println(test, ip, *ip)
	} else {
		fmt.Println(test, ip)
		os.Exit(1)
	}
}
