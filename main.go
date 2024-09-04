package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func main() {
    config := read_config("config.yasgp")
    for pattern, g := range config.groups {
        if pattern == "" { pattern = "/" }
        http.HandleFunc(pattern, func(w http.ResponseWriter, req *http.Request) {
            log.Printf("%v", req.URL.String())

            hostname := req.URL.Hostname()
            for _, rule := range g {
                if  hostname == "" { hostname = strings.Split(req.Host, ":")[0] }
                if hostname == rule.trigger.Host {
                    target_url := req.URL
                    target_url.Scheme = "http"
                    target_url.Host = rule.target.Host
                    joined_path, err := url.JoinPath(rule.target.Path, req.URL.Path)
                    if err != nil { log.Fatalln(err) }
                    target_url.Path = joined_path 
                    newReq, err := http.NewRequest(req.Method, target_url.String(), req.Body)
                    if err != nil {
                        http.Error(w, err.Error(), http.StatusInternalServerError)
                        return
                    }
                    // Copy the headers from the original request
                    for key, val := range req.Header {
                        newReq.Header[key] = val
                    }

                    newReq.Header.Set("x-forwarded-host", req.Host)
                    // Send the request via a http.Client
                    client := &http.Client{
                        CheckRedirect: func(req *http.Request, via []*http.Request) error {
                            return http.ErrUseLastResponse
                        },
                    }
                    res, err := client.Do(newReq)
                    if err != nil {
                        http.Error(w, err.Error(), http.StatusBadGateway)
                        return
                    }
                    // Copy the response headers and body back to the original writer
                    for key, val := range res.Header {
                        w.Header()[key] = val
                    }
                    w.WriteHeader(res.StatusCode)
                    io.Copy(w, res.Body)
                    return
                }
                http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
                return
            }
            http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
            return
        })
    }
    if err := http.ListenAndServe(fmt.Sprintf(":%v", config.port), nil); err != nil {
        log.Fatalf("Could not start server: %s\n", err.Error())
    }
}

type Rule struct {
    trigger url.URL
    target url.URL
}

func parse_rule(rule_string string) (*Rule, error) {
    rule_components := strings.Split(rule_string, " to ")
    if len(rule_components) < 2 { return nil, fmt.Errorf("rule parse error") }
    trigger, err := url.Parse(rule_components[0])
    if err != nil { return nil, err }
    target, err := url.Parse(rule_components[1])
    if err != nil { return nil, err }
    return &Rule{
        trigger: *trigger,
        target: *target,
    }, nil
}
type Config struct {
    port uint16
    rules []Rule
    groups map[string][]*Rule
}
func parse_config(config_file_data string) Config {
    config_file_lines := strings.Split(config_file_data, "\n")

    var port uint16
    port = 80
    rules := make([]Rule, 0)
    groups := make(map[string][]*Rule, 0)

    for i, r := range config_file_lines {
        if strings.Contains(r, "port") {
            p, err := strconv.ParseInt(strings.Split(r, " ")[1], 10, 32)
            if err != nil {
                panic(err)
            }
            port = uint16(p) 
        } else if r != "" {
            rule, err := parse_rule(r)
            if err != nil {
                log.Fatalf("%s at line %v\n", err.Error(), i)
            }
            rules = append(rules, *rule) 
            if g, ok := groups[rule.trigger.Path]; ok {
                groups[rule.trigger.Path] = append(g, rule)
            } else if groups[rule.trigger.Path] == nil {
                groups[rule.trigger.Path] = []*Rule{ rule } 
            }
        }
    }
    return Config{
        port: port,
        rules: rules,
        groups: groups,
    }
}

func read_config(filename string) Config {
    if !fs.ValidPath(filename) {
    }
    file_data, err := os.ReadFile(filename)
    if err != nil {
        os.Stderr.WriteString(err.Error())
    }
    file_data_str := string(file_data)
    return parse_config(file_data_str) 
}
