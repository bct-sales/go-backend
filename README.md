# Overview

## Installation

```bash
# Download file
$ wget https://github.com/bct-sales/go-backend/releases/latest/download/bctbackend

# Make it executable
$ chmod 766 ./bctbackend

# Create config file
$ ./bctbackend init

# Download HTML
$ ./bctbackend download

# Download font
$ wget https://github.com/googlefonts/noto-fonts/raw/main/hinted/ttf/NotoSans/NotoSans-Regular.ttf -O noto.ttf

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
