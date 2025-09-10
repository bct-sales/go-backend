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

* Generates a configuration file (does not overwrite existing)
* Downloads the latest HTML file (does not overwrite existing)
* Downloads `noto.ttf` (does not overwrite existing)
* Creates new database and populates it with default categories (does not overwrite or modify existing)

Note: `noto.ttf` lacks the arrow symbol used on labels.
Arial is a better choice.

### HTML Only

```bash
$ ./bctbackend init --html-only
```

Downloads the latest HTML file.
Overwrites existing HTML.

## Category

### Listing

```bash
$ ./bctbackend [--no-color] category list
ID | Name
1  | Clothing 0-3 mos (50-56)
2  | Clothing 3-6 mos (56-62)
3  | Clothing 6-12 mos (68-80)
4  | Clothing 12-24 mos (86-92)
5  | Clothing 2-3 yrs (92-98)
6  | Clothing 4-6 yrs (104-116)
7  | Clothing 7-8 yrs (122-128)
8  | Clothing 9-10 yrs (128-140)
9  | Clothing 11-12 yrs (140-152)
10 | Books
11 | Toys
12 | Baby/Child Equipment
13 | Maternity
```

Lists categories.

Note: `[--no-color]` means that `--no-color` is optional:

```bash
$ ./bctbackend category list

# or

$ ./bctbackend --no-color category list
```

### Item Counts by Category

```bash
$ ./bctbackend category count
ID | Name                         | Count
1  | Clothing 0-3 mos (50-56)     | 60
2  | Clothing 3-6 mos (56-62)     | 64
3  | Clothing 6-12 mos (68-80)    | 79
4  | Clothing 12-24 mos (86-92)   | 77
5  | Clothing 2-3 yrs (92-98)     | 67
6  | Clothing 4-6 yrs (104-116)   | 69
7  | Clothing 7-8 yrs (122-128)   | 59
8  | Clothing 9-10 yrs (128-140)  | 70
9  | Clothing 11-12 yrs (140-152) | 77
10 | Books                        | 589
11 | Toys                         | 568
12 | Baby/Child Equipment         | 0
13 | Maternity                    | 0
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
ID   | Description                   | Price | Category                     | Seller | Donation | Charity | Added At                       | Frozen | Hidden
------------------------------------------------------------------------------------------------------------------------------------------------------------
1    | purple scarf                  | 15.00 | Clothing 9-10 yrs (128-140)  | 100    | false    | false   | 2025-07-29 14:26:57 +0200 CEST | true   | false
2    | Nintendo Wii U                | 1.50  | Toys                         | 100    | false    | false   | 2025-08-01 00:11:54 +0200 CEST | true   | false
3    | green socks                   | 22.50 | Clothing 7-8 yrs (122-128)   | 100    | false    | false   | 2025-08-05 22:03:54 +0200 CEST | true   | false
4    | Finnegans Wake                | 35.00 | Books                        | 100    | false    | false   | 2025-08-08 09:43:32 +0200 CEST | true   | false
5    | red sweater                   | 24.00 | Clothing 11-12 yrs (140-152) | 100    | false    | true    | 2025-08-17 22:58:12 +0200 CEST | true   | false
...

# In CSV format
$ ./bctbackend item list --format csv
item_id,seller_id,description,category,price_in_cents,donation,charity
1,100,purple scarf,Clothing 9-10 yrs (128-140),1500,false,false
2,100,Nintendo Wii U,Toys,150,false,false
3,100,green socks,Clothing 7-8 yrs (122-128),2250,false,false
4,100,Finnegans Wake,Books,3500,false,false
5,100,red sweater,Clothing 11-12 yrs (140-152),2400,false,true
...
```

### Viewing an Item

```bash
$ ./bctbackend [--no-color] item show <id>

# Example
$ ./bctbackend --no-color item show 1
Property    | Value
--------------------------------------------
Description | purple scarf
Price       | 15.00
Category    | Clothing 9-10 yrs (128-140)
Seller      | 100
Donation    | false
Charity     | false
Added At    | 2025-07-29 14:26:57 +0200 CEST
Hidden      | false
Frozen      | true
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
* Either all items are updated, or none are, i.e., one frozen item in the list will block all other items from being frozen.

### Freezing/Unfreezing

```bash
$ ./bctbackend item [freeze|unfreeze] <id> ...
```

* A hidden item cannot be frozen.
* A sold item can be unfrozen (not recommended).
* Either all items are updated, or none are.

Unfreeze with caution!
Unfreezing allows editing the item, which may cause discrepancies with printed labels.

### Removing an Item

```bash
$ ./bctbackend item remove <id>
```

This command exists primarily for debugging purposes.
Only use this command in production if *really* necessary.
Hiding item is much safer.

## Database

### Backup

```bash
$ ./bctbackend db backup <backup_filename>
```

### Dummy Data

```bash
$ ./bctbackend db dummy [--seed <uint>]
```

Use with extreme caution: overwrites all data in the database.

## Users

### List All Users

```bash
$ ./bctbackend [--no-color] user list
```

### View User Information

```bash
$ ./bctbackend [--no-color] user show <id>
```

### Adding User

```bash
$ ./bctbackend user add --id <id> --password <password> --role <seller|cashier|admin>
```

### Add Sellers

```bash
$ ./bctbackend user add-sellers --zones <string>  --per-zone <int> --seed <uint>
```

Helper command to quickly add many sellers in one go.

* Zone can be enumeration of numbers: `--zones 1,2,3`.
* Zone can be range: `--zones 1-12`.
* Zone can be combination of enumeration and range.
* Seed determines how passwords are assigned to new sellers.
* Sellers will be assigned IDs in range `Z*100+K` where `Z` is zone number and `K` range from `0` to `--per-zone`.
* Existing sellers will be preserved.
* This commands adds sellers so that there are `--per-zone` sellers per zone listed; it does not necessarily add that many sellers.

### Changing Password

```bash
$ ./bctbackend user set-password <id> <password>
```

### Removing a User

```bash
$ ./bctbackend user remove <id>
```

Should never be used.

## Sales

### Listing Sales

```bash
$ ./bctbackend [--no-color] sale list
```

### Viewing Sales Details

```bash
$ ./bctbackend [--no-color] sale show <id>
```

### Add Sale

```bash
$ ./bctbackend [--no-color] sale add --cashier <id> --items <id>,<id>,...
```

The same item cannot appear multiple times in the same sale, but can appear in multiple different sales.

### Remove All Sales

```bash
$ ./bctbackend [--no-color] sale remove-all
```

Should never be used.
