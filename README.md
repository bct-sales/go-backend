# Overview

## Installation

```bash
# Download file
$ wget -O bctbackend https://github.com/bct-sales/go-backend/releases/latest/download/bctbackend

# Make it executable
$ chmod u+x ./bctbackend

# Initialize
$ ./bctbackend init

# Check configuration files
$ sudo nginx -t

# Restart nginx
$ sudo systemctl reload nginx
```

## Nginx Configuration

File `/etc/nginx/sites-available`

```text
server {
    listen 80;
    listen [::]:80;
    server_name bct-sales.myaddr.io 18.199.89.248;

    location /files/ {
        alias /var/www/html/;
    }


    location / {
        proxy_pass         http://127.0.0.1:8000;
        proxy_http_version 1.1;
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
    }
}
```

## DNS

Needs to happen monthly.

```bash
$ curl -d key=KEY -d ip=IP https://myaddr.tools/update
```

## Command Line Options

```bash
$ bctbackend user list
$ bctbackend item list
```

## Swagger

`http://localhost:8000/swagger/index.html`
