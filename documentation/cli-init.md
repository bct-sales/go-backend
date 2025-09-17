# CLI Initialization Commands

## Full Initialize

```bash
$ ./bctbackend init
```

Performs the following tasks:

* Generates a configuration file (does not overwrite existing)
* Downloads the latest HTML file (does not overwrite existing)
* Downloads `noto.ttf` (does not overwrite existing)
* Creates new database and populates it with default categories (does not overwrite or modify existing)

Note: `noto.ttf` lacks the arrow symbol used on labels.
Arial is a better choice.

## HTML Only

```bash
$ ./bctbackend init --html-only
```

Downloads the latest HTML file.
Overwrites existing HTML.
