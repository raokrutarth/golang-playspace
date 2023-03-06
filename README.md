# golang playspace

Playspace for golang expriments, code snippets and go tooling.

## Ideas

- render/serve/query a poorly served dataset
  - real-estate
    - trees per city ranking.
  - stock analysis
  - warehouse/shipping data
  - gov report data
  - people correctness scores.
- anonymous pool message board.
  - might need research for a secure solution.
  - needs to forget client connection identifiers like IP.
  - no identities/usernames. RO content board.
- timer file share app.

## Requirements

- authn with idp server
- uses htmx for simple UI.
- has open source telemetry built in with prom, jager & loki

## TODOs

- add cookie for session id, username.
- add pg store implementation CRUD.
- define template modules for summary, chart, expanded.

```
go run github.com/raokrutarth/golang-playspace/cmd/withauth
docker exec -it gops-dev-db env
docker exec -it gops-dev-db psql -U app -d gops_db -c "\dS public.*"

psql -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
psql -c "select * from range_transactions"

```

```
                    <!-- <div class="input-field col s5">
                        <input placeholder="{{.Prefill.Now}}" id="simulation_start" type="text" class="datepicker">
                        <label for="simulation_start">Simulation Start</label>
                    </div> -->
```
