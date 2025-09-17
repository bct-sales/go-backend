# CLI Item Commands

## Listing Items

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

## Viewing an Item

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

## Adding Item

```bash
$ ./bctbackend item add --category <id> -description <description> --price <price> --seller <seller_id> [--charity] [--donation]
```

A newly added item is always unfrozen and visible.

## Copying an Item

```bash
$ ./bctbackend item copy <id>
```

A copy of an item is always unfrozen and visible.

## Updating an Item

```bash
$ ./bctbackend item update --id <item_id> [--category <category_id>] [--charity] [--no-charity] [--donation] [--no-donation] [--price <price>] [--description <description>]
```

Cannot update a frozen item.

## Hiding/Unhiding

```bash
$ ./bctbackend item [hide|unhide] <id> ...
```

* A frozen item cannot be hidden.
* A sold item (which was unfrozen) can be hidden (not recommended).
* Either all items are updated, or none are, i.e., one frozen item in the list will block all other items from being frozen.

## Freezing/Unfreezing

```bash
$ ./bctbackend item [freeze|unfreeze] <id> ...
```

* A hidden item cannot be frozen.
* A sold item can be unfrozen (not recommended).
* Either all items are updated, or none are.

Unfreeze with caution!
Unfreezing allows editing the item, which may cause discrepancies with printed labels.

## Removing an Item

```bash
$ ./bctbackend item remove <id>
```

This command exists primarily for debugging purposes.
Only use this command in production if *really* necessary.
Hiding item is much safer.
