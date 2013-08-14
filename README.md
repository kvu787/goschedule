# Warning

[Nasty code ahead.](http://theprofoundprogrammer.com/post/31404285177/text-debugging-will-continue-until-morale)

# Deployment

- `go get github.com/kvu787/go-schedule/scraper`
- `go get github.com/kvu787/go-schedule/web`
- Create the databases specified in scraper/config/database.go
- In scraper/utility/sql
    - Run schema.sql against the two app databases and test database
    - Run switch.sql against the switch database
- Run `scraper` at least once to fill up an app database
- Run `web`
- Pray 
- Visit localhost:8080

# Setting up Postgres

- Edit `/etc/postgresql/9.2/main/pg_hba.conf` and change authentication to `md5`
- `sudo su postgres`
- `psql` -> create default role/database
- Login to default role
- Create application user/databases

# Nginx config

In /usr/local/nginx/conf

```
worker_processes  1;

events {
    worker_connections  1024;
}

http {
    include       mime.types;
    default_type  application/octet-stream;
    sendfile        on;
    keepalive_timeout  65;

    server {
        listen 80;
        server_name localhost;
        root <path to the directory where web command was run>;
        index index.html;

        location / {
                 try_files $uri $uri/;
        }

        location ~ /.* {
                include         fastcgi.conf;
                fastcgi_pass    127.0.0.1:9000;
        }

        try_files $uri $uri.html =404;
    }
}
```