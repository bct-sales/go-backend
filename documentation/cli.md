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
$ ./bctbackend --id <id> --name <name>

# Example

$ ./bctbackend --id 14 --name "Large Items"
```

No other category with same ID or name must exist, otherwise error.

## Items

### Listing Items

```bash
# In tabular form
$ ./bctbackend item list

# In CSV format
$ ./bctbackend item list --format csv
```

### Viewing an Item

```bash
$ ./bctbackend [--no-color] item show <id>
```

### Adding Item

```bash
$ ./bctbackend item add --category <id> -description <description> --price <price> --seller <seller_id> [--charity] [--donation]
```

A newly added item is always unfrozen and visible.

### Copying an Item

```bash
$ ./bctbackend item copy <id>
```

A copy of an item is always unfrozen and visible.

### Updating an Item

```bash
$ ./bctbackend item update --id <item_id> [--category <category_id>] [--charity] [--no-charity] [--donation] [--no-donation] [--price <price>] [--description <description>]
```

Cannot update a frozen item.

### Hiding/Unhiding

```bash
$ ./bctbackend item [hide|unhide] <id> ...
```

* A frozen item cannot be hidden.
* A sold item (which was unfrozen) can be hidden (not recommended).

### Freezing/Unfreezing

```bash
$ ./bctbackend item [freeze|unfreeze] <id> ...
```

* A hidden item cannot be frozen.
* A sold item can be unfrozen (not recommended).

Unfreeze with caution!
Unfreezing allows editing the item, which may cause discrepancies with printed labels.

### Removing an Item

```bash
$ ./bctbackend item remove <id>
```

Only use this command when *really* necessary.
Hiding item is much safer.
