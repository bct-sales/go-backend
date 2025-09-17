# CLI User Commands

## List All Users

```bash
$ ./bctbackend [--no-color] user list
```

## View User Information

```bash
$ ./bctbackend [--no-color] user show <id>
```

## Adding User

```bash
$ ./bctbackend user add --id <id> --password <password> --role <seller|cashier|admin>
```

## Add Sellers

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

## Changing Password

```bash
$ ./bctbackend user set-password <id> <password>
```

## Removing a User

```bash
$ ./bctbackend user remove <id>
```

Should never be used.

## Moving Items

```bash
$ ./bctbackend user move-items --from <id1> --to <id2> [--force-frozen] [--force-merge]
```

All items belonging to seller `id1` are reassigned to seller `id2`.

* If one or more of seller `id1`'s items is frozen, the operation fails, unless `--force-frozen` is specified.
* If seller `id2` already owns items, the operation fails, unless `--force-merge` is specified.
