# GitHub Traffic Logger
This tool is written in Go. It reads Insights/Traffic data for any GitHub repo 
and saves it in csv file.

The reasoning is that Traffic on holds up to 14 days of data, and has no option
to provide long-term statistics. It logs both number of repos Clones (all and
uniques) as well as Views (all and uniques).

This utility reads data from Traffic, and appends it to csv file, for easy
long-term monitoring.

It's suggested to run this tool once a week, not to miss any days.

## Building
```
go build ghtlogger.go
```

## Running
usage is pretty simple: `ghtlogger <owener/repo> -t path/to/gh_token -c path/to/csv`
for example:
```
ghtlogger jakubmatyszczak/github_traffic_logger -t ~/gh_token -c ~/Documents/ghtlogger.csv
```

## Crontab
Tool can be automated with cron. There is a script for automated crontab entry 
generation, that can guide you thought the process. Simply run:
```
./create_crontab_entry.sh
```
and follow the instructions.



