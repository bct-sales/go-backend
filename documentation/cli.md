# Command Line Interface

* [Server](./cli-server.md)
* [Initialization](./cli-init.md)
* [Items](./cli-items.md)
* [Categories](./cli-categories.md)
* [Database](./cli-db.md)

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
