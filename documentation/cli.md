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

## Items

### Listing Items

```bash
# In tabular form
$ ./bctbackend item list

# In CSV format
$ ./bctbackend item list --format csv
```

### Adding Item

```bash
$ ./bctbackend item add --category CATEGORY_ID -description DESCRIPTION --price PRICE --seller SELLER_ID [--charity] [--donation]
```

A newly added item is always unfrozen and visible.

### Copying an Item

```bash
$ ./bctbackend item copy ITEM_ID
```

A copy of an item is always unfrozen and visible.

### Hiding/Unhiding

```bash
$ ./bctbackend item [hide|show] ITEM_ID
```

* A frozen item cannot be hidden.
* A sold item (which was unfrozen) can be hidden (not recommended).

### Freezing/Unfreezing

```bash
$ ./bctbackend item [freeze|unfreeze] ITEM_ID
```

* A hidden item cannot be frozen.
* A sold item can be unfrozen (not recommended).

Unfreeze with caution!
Unfreezing allows editing the item, which may cause discrepancies with printed labels.
