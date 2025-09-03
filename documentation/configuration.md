# Configuration File

Use `$ ./bctbackend init` to automatically generate a configuration file.

```yaml
database: "bct.db"         # Database file
labelGeneration:
  barcode:                 # Size of bar code on labels
    width: 150
    height: 25
  font:                    # Font to use on labels. Note: Arial is a better alternative
    directory: "."
    filename: "noto.ttf"
    family: "Noto"
server:
  port: 8000               # nginx takes care of routing 80 to 8000
  html: index.html
  pruneExpiredSessionsInterval: 3600  # seconds before expired sessions are removed from database; does not impact functionality, only database size
  cookieDomain: "bct-sales.myaddr.io"
  debug: false
  swagger: false
log:
  file: "bct.log"
  maxSizeMegabytes: 10
  maxBackups: 3
  maxAgeDays: 28
  compression: false
```

## Swagger

In case swagger is set to `true`, the URL is

`http://URL:8000/swagger/index.html`
