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
$ ./bctbackend [--no-color] sale add --cashier <id> --items <id>,<id>,...
```

The same item cannot appear multiple times in the same sale, but can appear in multiple different sales.

## Remove All Sales

```bash
$ ./bctbackend [--no-color] sale remove-all
```

Should never be used.
