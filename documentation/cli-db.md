# CLI Database Commands

## Backup

```bash
$ ./bctbackend db backup [backup_filename]
```

The `backup_filename` is optional.
If omitted, `backup-YYYY-mm-dd-hh-mm-ss.db` will be used.
Fails if a file with the backup's filename already exists.

## Dummy Data

```bash
$ ./bctbackend db dummy [--seed <uint>] --overwrite
```

Use with extreme caution: overwrites all data in the database.
