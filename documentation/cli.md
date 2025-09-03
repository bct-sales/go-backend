# Command Line Interface

## Starting Server

```bash
$ ./bctbackend server
```

## Initialize

### Full Initialize

```bash
$ ./bctbackend init
```

Performs the following tasks:

* Generates a configuration file (does not overwrite)
* Downloads the latest HTML file (does not overwrite)
* Downloads noto.ttf (does not overwrite)
* Creates new database and populates it with default categories

### HTML Only

```bash
$ ./bctbackend init --html-only
```

Downloads the latest HTML file.
Overwrites existing HTML.

## Category

### Listing

```bash
$ ./bctbackend category list

# In case terminal does not support ANSI color codes
$ ./bctbackend --no-color category list
```

Lists categories.

### Item Counts by Category

```bash
$ ./bctbackend category count
```

### Adding New Category

```bash
$ ./bctbackend --id ID --name NAME

# Example

$ ./bctbackend --id 14 --name "Large Items"
```

No other category with same ID or name must exist, otherwise error.
