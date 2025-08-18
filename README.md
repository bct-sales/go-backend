# Overview

## Installation

```bash
# Download file
$ wget https://github.com/bct-sales/go-backend/releases/latest/download/bctbackend

# Make it executable
$ chmod u+x ./bctbackend

# Initialize
$ ./bctbackend init

sudo nginx -t
sudo systemctl reload nginx
```

## Nginx Configuration

```
server {
    listen 80;
    listen [::]:80;
    server_name bct-sales.duckdns.org 18.199.89.248;

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

## Command Line Options

```bash
$ bctbackend user list
$ bctbackend item list
```

## Swagger

`http://localhost:8000/swagger/index.html`

## Use Cases

### Admin

* View items
* View sales
* View users
* Deactivate sale (in case payment failed)

### Seller

* View items
* Add item
* Edit item
  * Only unfrozen items should be editable
* Print labels
  * Which items should be selectable
  * Freezes items

### Cashier

* Create sale
