# Archive Service

## TODO

### Immediate Steps

- email categories.
  - vc emails
  - tech innovation
  - linkedin notifications.
  - personal family emails.
  - deal alerts
  - job alerts
  - craigslist sale & old messages
  - dmv, bmv

- removal regex
  - linkedin message, invite, connect
  - moocho promotions
  - fortune ceo, science daily, hispotion

- what is needed for keyword analysis?
- What is needed for NER and extraction?
- how are DB backups going to be maintained?
- restore unread status from pst file.

- questions to answer:
  - What are the keywords mentioned in job alerts?
  - What companies are mentioned in job alerts?
  - What are the products on deals in specific times?
    - At what prices?
    - Use NLP model?

- IMAP client API: <https://pkg.go.dev/github.com/emersion/go-imap@v1.2.1/client#Client>
- str-utils
  - <https://github.com/ozgio/strutil>

- restart development from:
  - <https://godocs.io/github.com/emersion/go-imap/client#example-Client-Search>
  - <https://gorm.io/docs/update.html>
  - <https://github.com/flavio912/react-golang-eLearning/blob/775c3396ae411973a5fe566387750bb64026232a/api/database/migration/migration.go>
  - <https://gorm.io/docs/has_one.html>
  - <https://gorm.io/docs/migration.html>
  - <https://medium.com/@hemanthponnada23/gmail-data-analysis-using-python-184cc4a8f35b>
  - <https://github.com/donomii/shonkr/blob/6261545c6d47c623fe6043fec2130f4129018e1f/v3/getmail.go>
  - <https://github.com/budenny/mail2telegram/blob/4b77de2b54ba482eccb907d40fa329f4757d8843/mail/client.go#L85>
  - <https://github.com/budenny/mail2telegram/blob/4b77de2b54ba482eccb907d40fa329f4757d8843/mail/message.go#L22>
  - <https://github.com/budenny/mail2telegram/blob/4b77de2b54ba482eccb907d40fa329f4757d8843/main.go>
  - <https://github.com/emersion/go-imap/wiki/Fetching-messages>
  - <https://github.com/thedustin/go-email-curator/blob/54c33f2d542d4c20e8a72fc03a0d88fc9d118253/action/move.go>

- search code samples:
  - <https://github.com/xmapst/alps/blob/master/plugins/base/search.go>
  - <https://github.com/emersion/go-imap/blob/master/backend/backendutil/search_test.go>
  - <https://github.com/emersion/go-imap/blob/master/search_test.go>
  - <https://github.com/stbenjam/go-imap-notmuch/blob/main/pkg/notmuch/search_test.go>

- logging
  - Log to file + stdout using zerolog.multiwriter and lumberjack.
  - sp <https://betterstack.com/community/guides/logging/zerolog/#logging-to-a-file>
  - <https://www.ribice.ba/go-logging/>
  - <https://gist.github.com/panta/2530672ca641d953ae452ecb5ef79d7d>
  - <https://dev.to/shiguredo/zerolog-lumberjack-io-multiwriter-2a7k>
  -
  - <https://pkg.go.dev/tawesoft.co.uk/go/log/zerolog#section-readme>
- message parsing
  - <https://git.sr.ht/~emersion/alps/tree/master/plugins/base/imap.go#L147> message parsing and structs

- cli lib & prompting
  - <https://divrhino.com/articles/build-interactive-cli-app-with-go-cobra-promptui/>

- testing
  - <https://github.com/stretchr/testify>

- ORM/DB
  - <https://gorm.io/docs/update.html>
  - json data: <https://gorm.io/docs/v2_release_note.html#DataTypes-JSON-as-example>

- EDA
  - clickhouse local
    - <https://clickhouse.com/blog/extracting-converting-querying-local-files-with-sql-clickhouse-local>
    - data output formats <https://clickhouse.com/docs/en/interfaces/formats/>
    - <https://clickhouse.com/blog/real-world-data-noaa-climate-data>
  - alternatives
    - <https://github.com/cube2222/octosql>
    - <https://github.com/dinedal/textql>
    - <https://github.com/BurntSushi/xsv>
    - <https://github.com/harelba/q>
    - <https://github.com/multiprocessio/dsq> for json output.
    - <https://github.com/liquidaty/zsv>
    - <https://github.com/noborus/trdsql>
- search email code
  - <https://git.sr.ht/~emersion/alps/tree/master/plugins/base/imap.go> attachments
  - <https://github.com/rodrigol-chan/aerc/blob/master/worker/imap/search.go>
  - <https://github.com/emersion/go-imap/blob/master/search_test.go>
  - <https://github.com/emersion/go-imap/blob/master/backend/backendutil/search_test.go>
  - <https://github.com/emersion/go-imap/blob/master/backend/backendutil/search.go>
  - <https://github.com/emersion/go-imap/blob/master/commands/search.go>
  - <https://github.com/foxcpp/go-imap-backend-tests/blob/master/mailbox_search.go>
  - <https://github.com/xmapst/alps/blob/master/plugins/base/search.go> search OR, AND utils
  - <https://github.com/stbenjam/go-imap-notmuch/blob/main/pkg/notmuch/search_test.go>
  - <https://github.com/foxcpp/go-imap-sql/blob/c20be1a387b4bc727cda5ed4079996d87c7c89e1/search.go> flag search
  - <https://pkg.go.dev/github.com/emersion/go-imap?utm_source=godoc#SearchCriteria> search godoc

```txt
Received:01/01/2017..01/01/2019
```

## Functionality

- Read email using smtp/imap.
- clean/transform as necessary.
- index it into the search layer.

## Ideas

- Testing
  - use mailhog container with test account and bootstrap data.

## Resources

### Storage Engine

- <https://gitea.com/a1012112796/test_go_imap/src/branch/master/imap.go> delete and mark read code.
- `https://github.com/typesense/typesense#api-clients`: Typesense is a fast, typo-tolerant search engine for building delightful search experiences.
- `https://github.com/valeriansaliou/sonic`: Rust based search system.
- `https://github.com/groonga/groonga` written in C.
- `https://crate.io/products/cratedb` good for metric and text store.
- `https://github.com/meilisearch/meilisearch` Rust but not yet v1
- `https://github.com/manticoresoftware/manticoresearch/` faster than ES. written in C++
- `https://github.com/qdrant/qdrant` rust based
- `https://pyvespa.readthedocs.io/en/latest/getting-started-pyvespa.html` not clear

## dev commands

- `migrate -source file://platform/migrations -database 'postgres://dev:djZMi4hGgSLpbc1B@db:5432/cashflow?sslmode=disable' up`
- `psql postgres://dev:djZMi4hGgSLpbc1B@db:5432/cashflow`
- `migrate create -ext sql -dir platform/migrations -seq add_user_json_col`
- Download prod DB cert `curl --create-dirs -o $HOME/.postgresql/root.crt -O https://cockroachlabs.cloud/clusters/b44f6363-34a1-4935-be55-41fae9623fa6/cert`

## Docs

- Postgress migration operations `https://www.postgresql.org/docs/12/ddl.html`
- Postgres common migrate ops `hhttps://www.postgresqltutorial.com/`
- cockroachdb URI params `https://www.cockroachlabs.com/docs/v22.1/connection-parameters#additional-connection-parameters`

### Core

- `https://madflojo.medium.com/how-to-structure-a-golang-project-aad7095d70a` mvp main.go file
- `https://github.com/reugn/go-quartz` in-memory scheduler.
- `http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/` worker pool.
- `https://github.com/hashicorp/terraform/blob/main/Makefile` go project layout and makefile
- `https://github.com/hbollon/IGopher/blob/master/internal/config/config.go` config reading
- `https://madflojo.medium.com/using-viper-with-consul-to-configure-golang-applications-eaa84394b8de` config lib
- `https://pkg.go.dev/github.com/PuerkitoBio/goquery` html golang parser.
  - `https://github.com/Arnesh07/golang-python-web-scraping/blob/master/go_scraper/scraper_par_gocolly_parallelism.go` sample code.
- `https://golangcode.com/basic-web-scraper/` html parser snippet
- alternative go imap email client `https://pkg.go.dev/github.com/mxk/go-imap/imap`

### Infra

- `https://medium.com/scum-gazeta/golang-production-ready-solution-part-3-8c9d8d2835c6` dockerfile
