# golang playspace

Playspace for golang expriments, code snippets and go tooling.

## Usage

- spin up dev containers in VSCode <https://code.visualstudio.com/docs/containers/docker-compose>
- copy and fill out the `dev.env` with `cp dev.env.example dev.env`
- In the dev container, run `air` to start the BE.
- May need to setup ports using the VS code "Ports" tab.
- Open the BE on `localhost`.

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

## References

- boilerplate with authn
  - <https://github.com/karlkeefer/pngr>
- simple lists server and templates
  - <https://github.com/benhoyt/simplelists/blob/e3a7f93f1310d72b20bb7b47fb24c0f0930b79f4/server_test.go>
- gorm boilerplate with form parsing and validation
  - <https://github.com/learning-cloud-native-go/myapp/blob/b74a8391ee101de52db4e2590421aca45a97a1bf/api/resource/book/repository.go>
    - validation error parsing: <https://github.com/learning-cloud-native-go/myapp/blob/dc5beaf69250effa8207a2181a2f665858217c71/util/validator/validator.go>
- html templates
  - <https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/html-templates>
  - with file embed: <https://charly3pins.dev/blog/learn-how-to-use-the-embed-package-in-go-by-building-a-web-page-easily/>
  - <https://www.calhoun.io/intro-to-templates-p2-actions/>
  - <https://www.calhoun.io/intro-to-templates-p3-functions/>
- struct validation
  - <https://thedevelopercafe.com/articles/payload-validation-in-go-with-validator-626594a58cf6>
  - <https://pkg.go.dev/github.com/go-playground/validator/v10#section-readme>
  - in fiber framework: <https://dev.to/franciscomendes10866/how-to-validate-data-in-golang-1f87>
  - iso dates with ech framework: <https://rickyanto.com/how-to-create-custom-validator-for-iso8601-date-with-go-validator-from-go-playground-and-echo-framework/>
- HTTP forms parsing lib: <https://github.com/go-playground/form>
- html charts
  - <https://canvasjs.com/docs/charts/chart-types/html5-line-chart/>
  - <https://www.w3schools.com/js/js_graphics_chartjs.asp>
  - <https://developers.google.com/chart/interactive/docs/quick_start>

## TODOs

- add cookie for session id, username.
- add pg store implementation CRUD.
- define template modules for summary, chart, expanded.

## Common Commands

```bash
go run github.com/raokrutarth/golang-playspace/cmd/withauth -add-user
docker exec -it gops-dev-db env

psql -c "\dS public.*"
psql -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
psql -c "TRUNCATE TABLE range_transactions, expanded_transactions CASCADE;"
psql -c "select * from range_transactions"
```
