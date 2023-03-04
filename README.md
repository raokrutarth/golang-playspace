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

- move to single file server with login: <https://github.com/benhoyt/simplelists/blob/master/server.go>
  - has csrf, delete button + confirm, sql init, insert & update queries,

```
go run github.com/raokrutarth/golang-playspace/cmd/withauth
docker exec -it gops-dev-db env
docker exec -it gops-dev-db psql -U app -d gops_db
```
