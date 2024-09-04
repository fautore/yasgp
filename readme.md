# YASGP - Yet Another Silly Go Proxy
YASGP is, as it's name implies, a proxy written in pure go

# usage
- write your configuration file to `config.yasgp` (currently yasgp looks for `config.yasgp` in the project root), more on that later
- compile and run
- enjoy your proxy!
# configuration
yasgp uses a configuration file separated in lines, below an example:
```yasgp
port 8080
http://foo.org to http://localhost:5000
http://foo.net to http://localhost:5001
http://bar.com/bar to http://localhost:5002
```
