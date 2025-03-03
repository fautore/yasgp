package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const VERSION = "0.0.1"

type Rule struct {
	trigger url.URL
	target  url.URL
}
type Address string

type Config struct {
	rules  []Rule
	groups GroupMap
}
type GroupMap struct {
	groups map[Address]*Group
}

func NewGroupMap() GroupMap {
	return GroupMap{
		groups: make(map[Address]*Group),
	}
}

func (g *GroupMap) has(address Address) bool {
	_, ok := g.groups[address]
	return ok
}
func (g *GroupMap) get(address Address) (*Group, error) {
	if !g.has(address) {
		return nil, fmt.Errorf("%v is not found in group", address)
	}
	group := g.groups[address]
	return group, nil
}
func (gm *GroupMap) set(address Address, group Group) {
	gm.groups[address] = &group
}

type Group struct {
	address Address
	rules   []Rule
}

func startProxy(group Group) {
	mux := http.NewServeMux()
	for _, rule := range group.rules {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			r.URL.Host = ""
			target_url, err := url.Parse(rule.target.String() + r.URL.String())
			if err != nil {
				log.Printf("%v", err.Error())
				return
			}

			log.Printf("Received request on binding %v, redirecting to %v ...\n", r.URL, target_url)

			header := make(http.Header)
			for hk, h := range r.Header {
				for _, hv := range h {
					header.Add(hk, hv)
				}
			}
			// log.Printf("richiesta:\n\n\n%v\n\n\n", header)

			res, err := http.DefaultClient.Do(&http.Request{
				Method: r.Method,
				URL:    target_url,
				Header: header,
			})
			if err != nil {
				log.Printf("%v", err.Error())
			}
			defer res.Body.Close()

			for hk, h := range res.Header {
				for _, hv := range h {
					w.Header().Add(hk, hv)
				}
			}

			_, err = io.Copy(w, res.Body)
			if err != nil {
				log.Printf("%v", err.Error())
			}
			w.WriteHeader(res.StatusCode)
		})
	}

	server := &http.Server{
		Addr:    string(group.address),
		Handler: mux,
	}

	log.Printf("Listening on address %v", group.address)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("%v\n", err.Error())
	}
}

func main() {
	fmt.Printf("YASGP - Yet Another Silly Go Proxy\nv.%v\n\n\n", VERSION)

	config, err := read_config_file("config.yasgp")
	if err != nil {
		log.Fatalf("%v", err)
	}

	for _, group := range config.groups.groups {
		for _, r := range group.rules {
			log.Printf("%v -> %v\n", r.trigger.Host, r.target.Host)
		}
		go startProxy(*group)
	}
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
	log.Printf("Received stop signal, closing...")
}

func parse_line(rule_string string) (*Rule, error) {
	// ignore empty lines
	if rule_string == "" {
		return nil, nil
	}
	// ignore comments
	if rule_string[0] == '#' {
		return nil, nil
	}
	rule_components := strings.Split(rule_string, " to ")
	if len(rule_components) < 2 {
		return nil, fmt.Errorf("rule parse error")
	}
	trigger, err := url.Parse(rule_components[0])
	if err != nil {
		return nil, err
	}
	target, err := url.Parse(rule_components[1])
	if err != nil {
		return nil, err
	}
	return &Rule{
		trigger: *trigger,
		target:  *target,
	}, nil
}

func read_config_file(filename string) (*Config, error) {
	if !fs.ValidPath(filename) {
		return nil, fmt.Errorf("file name %v is not a valid path", filename)
	}
	file_data, err := os.ReadFile(filename)
	if err != nil {
		os.Stderr.WriteString(err.Error())
	}
	file_data_str := string(file_data)
	// windows support
	config_file_lines := strings.Split(file_data_str, "\n")
	for i, l := range config_file_lines {
		config_file_lines[i] = strings.TrimSuffix(l, "\r")
	}

	rules := make([]Rule, 0)

	for i, r := range config_file_lines {
		if r != "" {
			rule, err := parse_line(r)
			if err != nil {
				log.Fatalf("%s at line %v\n", err.Error(), i)
			}
			if rule != nil {
				rules = append(rules, *rule)
			}
		}
	}

	config := Config{
		rules:  rules,
		groups: NewGroupMap(),
	}

	for _, r := range config.rules {
		address := Address(r.trigger.Host)
		// TODO debug log.Printf("%v\n", address)
		if !config.groups.has(address) {
			config.groups.set(address, Group{
				address: address,
				rules:   make([]Rule, 0),
			})
		}
		group, err := config.groups.get(address)
		if err != nil {
			return nil, fmt.Errorf("")
		}
		group.rules = append(group.rules, r)
	}
	return &config, nil
}
