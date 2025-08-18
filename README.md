# Overview

## Installation

```bash
# Download file
$ wget https://github.com/bct-sales/go-backend/releases/latest/download/go-backend_1.3.0_linux_amd64.tar.gz

# Unpack file
$ tar xzvf go-backend_1.3.0_linux_amd64.tar.gz
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
