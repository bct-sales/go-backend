# Running Locally on Windows

Running the server requires four files:

* The backend executable, named `bctbackend.exe` (Windows) or `bctbackend` (Linux).
* The frontend HTML, named `index.html`.
* The database, named `bct.db`.
* The configuration file, `bctconfig.yaml`.

We now explain below how to create/download these files.

## Downloading the Backend

The backend executable is available on [GitHub](https://github.com/bct-sales/go-backend/releases).
Download `bctbackend.exe` and place it in a directory of your choice.

## The Other Three Files

`bctbackend.exe` takes care of providing the three other files.
In a shell, go to the directory containing `bctbackend.exe` and run

```bash
$ ./bctbackend.exe init all
```

It will

* Download the latest `index.html` (also hosted on GitHub)
* Create an empty database (containing the necessary tables, but no data)
* Create a configuration file with reasonable defaults

All three files will be placed in the same directory, where they are expected.

## Updating the Configuration File

In order for the software to run locally, a modification needs to be made to the configuration file.
Open `bctconfig.yaml` in a text editor and set `server.cookieDomain` to `"localhost"`.

```yaml
...
server:
  ...
  cookieDomain: "localhost"
  ...
...
```

## Running the Server

You can start the server using the following command:

```bash
$ ./bctbackend.exe server
```

Opening your browser and going to `http://localhost:8000` should show you the login page.
You won't be able to login however, as no users exist.

Note that running the shell will "block" your shell, i.e., you will not be able to enter new commands as long as the backend is busy.
Shutting down the server can be done using CTRL+C (maybe you'll have to insist and press it multiple times).

## Asking Help from the Backend

A useful thing to know is that you can always ask the backend for help.
Open a new shell in the same directory, and run

```bash
$ ./bctbackend.exe --help
BCT Backend Command Line Interface for managing items, users, and other resources.

Usage:
  bctbackend [command]

Available Commands:
  category          Manage categories
  completion        Generate the autocompletion script for the specified shell
  db                Performs database-level operations
  email             Performs email related operations
  generate-password Generates password
  help              Help about any command
  init              Initialize backend components
  item              Manage items
  sale              Manage sales
  server            Start server
  session           Manage sessions
  ui                Start UI
  user              Manage users
  version           Shows version

Flags:
      --config string   Path to the configuration file
      --db string       Path to the database file (default "./bct.db")
  -h, --help            help for bctbackend
      --no-color        Disable color output
  -v, --verbose         Enable verbose output

Use "bctbackend [command] --help" for more information about a command.
```

This shows a list of commands.
You can then zoom in on a command and ask for further help:

```bash
$ ./bctbackend.exe init --help
Commands to initialize different parts of the back-end.

Usage:
  bctbackend init [command]

Available Commands:
  all         Initializes all components
  html        Download html

[unimportant lines left out]
```

Adding `--help` will always prevent the command from doing anything, so it's safe to experiment with.

## Adding an Admin

All user functionality resides under the `user` command:

```bash
$ ./bctbackend.exe user --help

Commands to manage users in the BCT backend system.

Usage:
  bctbackend user [command]

Available Commands:
  add          Add a new user
  add-sellers  Add multiple sellers
  list         List all users
  move-items   Move items from one seller to another
  remove       Removes a user
  set-password Sets user password
  show         Show user info
```

The `add` subcommand seems promising:

```bash
$ ./bctbackend.exe user add --help

This command adds a new user to the database.

Usage:
  bctbackend user add [flags]

Flags:
      --generate-password   Generate password for the user
  -h, --help                help for add
      --id int              ID of the user to add
      --password string     Password for the user
      --role string         Role of the user (admin, seller, cashier)
```

This tells us we can use

```
$ ./bctbackend.exe user add --id 1 --password admin --role admin
```

This adds a user with id 1, password set to `admin`, and is given the role of admin.
If you go back to your browser, you should be able to log in using `1` and `admin` as login and password.

## Adding a Seller

Similary, we can add a seller:

```bash
$ ./bctbackend.exe user add --id 100 --password abc --role seller
```

In the browser, you should be able to see the user appear in the user list.
You should also be able to log in as a seller using `100` and `abc`.

If you don't want to need to come up with a password, you can also write

```bash
$ ./bctbackend.exe user add --id 101 --generate-password --role seller
```

This will pick a random as-of-yet-unused password from a built-in dictionary of five letter words.

## Adding Many Sellers at Once

Typically you will want to create many sellers, spread across multiple zones.
A specialized command has been provided for this specific use case:

```bash
$ ./bctbackend.exe user add-sellers --zones 1-12 --per-zone 5
```

This command will create create sellers `100`, `101`, `102`, `103`, `104`, `200`...`204`, `300`...`304`, ..., `1200`...`1204`, each with a unique five letter word password.

If it turns out you did not add sufficient sellers, you can always call this command again with higher numbers.
Note that it will only create sellers so that in the end, you end up with the requested amount of zones containing the requested number of sellers.
More specifically, running the same command again

```bash
$ ./bctbackend.exe user add-sellers --zones 1-12 --per-zone 5
```

will do nothing, as there are already 12 zones with 5 sellers each.
If you call

```bash
$ ./bctbackend.exe user add-sellers --zones 1-13 --per-zone 6
```

This will create sellers `105`, `205`, `305`, ..., `1205`, `1300`...`1305`.
In other words, the command is safe to use to raise the number of zones/sellers up to a certain point.

## Listing Users from the Shell

While you can list all users from the web interface, you can also see them in the shell:

```bash
$ ./bctbackend.exe user list
```

Using `--help`, you can find out that it also allows you to specify a format:

```bash
$ ./bctbackend.exe user list --format csv
```

will print the same data out in CSV format.
Writing this output to file can be done using redirection:

```bash
# Write csv data to users.csv
$ ./bctbackend.exe user list --format csv > users.csv
```

## Backing Up the Database

The entire database resides in a single file named `bct.db`.
If the server is inactive, the database can be fully copied simply by copying the file.
However, if the server is active, it is important to use the following command:

```bash
$ ./bctbackend.exe db backup bct-backup.db
```

This creates a new database file named `bct-backup.db` to which all data is copied in a manner that guarantees consistency.

## Restoring a Backup

In order to go "back in time", the server needs to be stopped first (CTRL-C).
Next, overwrite `bct.db` with a database file of your choice.
Next, restart the server using `./bctbackend.exe server`.

## Adding Dummy Data

For testing purposes, a command exists to fill the database with dummy data.
Note that the dummy data will *replace* the existing contents of the database, so this command is quite dangerous.

```bash
$ ./bctbackend.exe db dummy --overwrite
```

## Other Commands

There are many other commands available, the most important ones being the item and sales ones:

```bash
$ ./bctbackend.exe item --help
Commands to manage items in the BCT backend system.

Usage:
  bctbackend item [command]

Available Commands:
  add            Add an item
  add-consumable Add a consumable item
  copy           Copies an item
  freeze         Freezes items
  hide           Hides items
  list           List all items
  remove         Remove items
  show           Show item info
  unfreeze       Unfreezes items
  unhide         Unhides a items
  update         Updates an item
```

```bash
$ ./bctbackend.exe sale --help
Commands to manage sales in the BCT backend system.

Usage:
  bctbackend sale [command]

Available Commands:
  add         Add a new sale
  items       List all sold items
  list        List all sales
  remove      Removes specific sales
  remove-all  Removes all sales
  show        Show a sale
```
