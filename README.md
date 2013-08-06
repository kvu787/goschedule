# Warning

[Nasty code ahead.](http://theprofoundprogrammer.com/post/31404285177/text-debugging-will-continue-until-morale)

# Deployment

- Install Nginx
- Verify Nginx server is working
- Install Go
    - Setup $GOPATH properly
- Verify Go works
- [Install PostgreSQL](http://www.postgresql.org/download/linux/ubuntu/)
- `go get github.com/kvu787/go-schedule`
- Create the databases specified in crawler/config/database.go
- In crawler/utility/sql
    - Run schema.sql against the two app databases
    - Run switch.sql against the swtich database
- Run crawler in background
- Update Nginx config (see Nginx config section) and restart Nginx server
- Run web in background
- Pray 
- Visit the hosting address to verify things work

# Running background processes (Unix)

- Run and do not exit on logout: `nohup <command> &`
- View processes and PIDs: `ps -e`
- Kill process: `kill <pid>`

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