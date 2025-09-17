# Command Line Interface

* [Server](./cli-server.md)
* [Initialization](./cli-init.md)
* [Items](./cli-items.md)


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
