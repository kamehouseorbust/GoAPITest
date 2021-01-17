# How to run:

    Download repo and GeoLite2-Country.mmdb, place db in same directory as main.go.

    Run:
    `go run main.go`

    Valid Query:
        - http://localhost:3000/api/ipWhiteListed?ip=<publicip>&whiteListValues=US&whiteListValues=FI&whiteListValues=KE
        - http://localhost:3000/api/ipWhiteListed?ip=<publicip>&whiteListValues=UK&whiteListValues=FI&whiteListValues=KE
    
    Invalid (For testing):
        - http://localhost:3000/api/ipWhiteListed?ip=1.3
        - http://localhost:3000/api/ipWhiteListed?ip=<publicip>

---

# TODO:

    - Write cronjob to replace database file once per week, would require your own maxmind account.
        `gocron.Every(1).Week().Do(refreshDBFile)`
**Note:** We'd probably need to restart the API, but downtime would only be a couple seconds.

    - Separate REST logic from main and add gRPC as a sibling package, both run from main.go
    - Separate constants into their own package so they can be used by main, REST, and gRPC.
    - Write docker file and kubernetes YAML for easy setup.
    - Need to switch to post for REST because we're probably going to run into a 414 error with the potential length of the whiteListValues param. This would also open the door for sending more than one ip at a time (we're currently only counting on one ip parameter)