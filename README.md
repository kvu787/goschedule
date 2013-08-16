# Deployment

- `go get github.com/kvu787/go-schedule/scraper`
    - NOTE: You will receive several "...undefined: config.*" errors. This is normal.
    - In `scraper/config/`, copy `config.go.example` to `config.go` and edit the settings as necessary.
- In `scraper/`, run `go install`.
- `go get github.com/kvu787/go-schedule/web`
- Create the databases specified in `scraper/config/database.go`.
- In `scraper/utility/sql`
    - Run schema.sql against the two app databases and test database.
    - Run switch.sql against the switch database.
- Run `scraper` at least once to fill up an app database.
- Run `web`.
- Pray.
- Visit localhost:8080.