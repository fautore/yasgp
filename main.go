package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func generateRequestHandler(target string) func(http.ResponseWriter, *http.Request) {
    return func(res http.ResponseWriter, req *http.Request) {
        fmt.Printf("req: \n%v\n", req)
        
        url := req.URL
        url.Scheme = "http"
        url.Host = target

        newReq, err := http.NewRequest(req.Method, url.String(), req.Body)
        if err != nil {
            http.Error(res, err.Error(), http.StatusInternalServerError)
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
        resp, err := client.Do(newReq)
        if err != nil {
            http.Error(res, err.Error(), http.StatusBadGateway)
            return
        }
        fmt.Printf("resp: \n%v\n\n", resp)

        // Copy the response headers and body back to the original writer
        for key, val := range resp.Header {
            res.Header()[key] = val
        }
        res.WriteHeader(resp.StatusCode)
        io.Copy(res, resp.Body)
    }
}

type ProgramArguments struct{
    program string
    port string
    target string
}
func parseArgs() ProgramArguments{
    var programArguments ProgramArguments
    programArguments.program = os.Args[0]
    arguments := os.Args[1:]

    for i := 0; i < len(arguments); i++ {
        arg := arguments[i]
        if arg == "-p" || arg == "--port" {
            if i+1 < len(arguments) {
                port := arguments[i+1]
                i = i+1
                programArguments.port = port
            }
        }
        if arg == "-t" || arg == "--target" {
            if i+1 < len(arguments) {
                target := arguments[i+1]
                i = i+1
                programArguments.target = target
            }
        }
    }
    if programArguments.port == ""{
        panic("Missing port argument")
    }
    if programArguments.target == "" {
        panic("Missing target argument")
    }
    return programArguments
}

func main() {
    args := parseArgs()
    http.HandleFunc("/", generateRequestHandler(args.target))
    fmt.Printf("yasgp running on port %s\n", args.port)
    if err := http.ListenAndServe(fmt.Sprintf(":%s", args.port), nil); err != nil {
        log.Fatalf("Could not start server: %s\n", err.Error())
    }
}
