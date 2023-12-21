# Cron

Cron is a simple library schedule functions to run periodically using 
[Unix Cron Format](https://www.ibm.com/docs/en/db2/11.5?topic=task-unix-cron-format).

A job (parameterless function) is scheduled with the following cron specification string:

```
  * * * * *
  | | | | |
  | | | | +- Day of week (0–7) (Sunday=0 or 7) or Sun, Mon, Tue,…
  | | | +--- Month (1–12) or Jan, Feb,…
  | | +----- Day of month (1–31)
  | +------- Hour (0–23)
  +--------- Minute (0–59)
```

Asterisk represents every single value within that field. Values, steps, lists and ranges are 
allowed within any field. Multiple values may be specified in a field, separate each with a comma.

* Values: 5,10,15
* Step: */12
* Range: 10-20
* Combine: 10-20,24,33

## Examples

* `* * * * *` run once per minute
* `*/12 * * * *` run at 0, 12, 24, 36, 48 minutes every hour
* `0 0 * * *` run at midnight
* `0 12 14 3 *` run at 12:30pm, on pie day (March 14)

## Command Line Example

Either print the time periodically or run a command:

```
go run ./cmd
go run ./cmd "*/3 * * * *"
go run ./cmd "*/5 * * * *" "ls -l"
```

## Use Cron within Go

```
package main

import (
    "fmt"
    "log"
	"os"
	"os/signal"
	"time"
	"github.com/sspencer/cron"
)

func main() {
    c, err := cron.Run("*/2 * * * *", func() {
        fmt.Printf("CRON: %s\n", time.Now().Format("15:04:05.000"))
    })

	if err != nil {
		log.Fatal(err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	c.Stop()
}
```
