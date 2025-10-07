# CLI Initialization Commands

## Full Initialize

```bash
$ ./bctbackend init all
```

Performs the following tasks:

* Generates a configuration file (does not overwrite existing)
* Downloads the latest HTML file (does not overwrite existing)
* Creates new database and populates it with default categories (does not overwrite or modify existing)

## HTML Only

```bash
$ ./bctbackend init html
```

Downloads the latest HTML file.
Overwrites existing HTML.
