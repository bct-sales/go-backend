# CLI Sale Commands

## Listing Sales

```bash
$ ./bctbackend [--no-color] sale list
```

## Viewing Sales Details

```bash
$ ./bctbackend [--no-color] sale show <id>
```

## Add Sale

```bash
$ ./bctbackend sale add --cashier <id> --items <id>,<id>,...
```

The same item cannot appear multiple times in the same sale, but can appear in multiple different sales.

## Remove Specific Sales

```bash
$ ./bctbackend sale remove <id1> <id2> ...
```

Removes sales with specified ids.
If removing one sale fails, then none of them are removed.

## Remove All Sales

```bash
$ ./bctbackend sale remove-all
```

Should never be used.
