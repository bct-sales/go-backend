# Installation

## Downloading Backend

```bash
# Download latest build from GitHub
$ wget -O bctbackend https://github.com/bct-sales/go-backend/releases/latest/download/bctbackend

# Make it executable
$ chmod u+x ./bctbackend

# Initialize
$ ./bctbackend init
```

## Setting Up Nginx (without TLS)

Create file `bct-backend` in `/etc/nginx-sites-available`:

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

```bash
# Create symbolic link in sites-enabled
$ cd /etc/nginx
$ ln -s sites-available/bct-backend sites-enabled

$ cd

# Check configuration files
$ sudo nginx -t

# Restart nginx
$ sudo systemctl reload nginx
```

## Adding TLS

```bash
$ sudo apt update
$ sudo apt install certbot python3-certbot-nginx
$ sudo certbot --nginx
```

This will cause `http://bct-sales.myaddr.io` to be redirected to `https://bct-sales.myaddr.io` and `https://bct-sales.myadd.io` to be server by the backend.
Note that the certificate needs to be renewed very 90 days, but certbot should register a systemd timer that takes care of the renewal.

```bash
$ systemctl list-timers | grep certbot
# Should show certbot.timer
```

## Starting backend

```bash
$ cd
$ nohup ./bctbackend server &
```

(Should be improved)

## Renewing DNS (!= HTTPS Certificate)

Needs to happen monthly.

```bash
$ curl -d key=KEY -d ip=self https://myaddr.tools/update
```
