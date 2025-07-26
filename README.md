# SBBuddy

## Current Functionalities
- Look up timetable of a station
  - Example: SBBuddy -T "Basel SBB"

- Search connections between two or more stations
  - Example: SBBuddy -C "Basel SBB" -C "Zürich HB"

- With via stations: SBBuddy -C "Basel SBB" -C "Olten" -C "Zürich HB"

- Fetch a random connection
  - Basic: SBBuddy -R
  - With number of via stations: SBBuddy -R 2
  - Combine with fixed station(s): SBBuddy -C "Basel SBB" -R, SBBuddy -R -C "Zürich HB"

- Timetable from a specific time and/or date
  - Examples:
  - SBBuddy -T "Basel SBB" -t 10:00
  - SBBuddy -T "Basel SBB" -d 2025-07-28 -t 14:50

- Connections from a specific date and/or time
  - Examples:
  - SBBuddy -C "Basel SBB" -C "Zürich HB" -d 2025-08-14
  - SBBuddy -C "Basel SBB" -C "Zürich HB" -d 2025-08-14 -t 10:00

- Scroll through connection lists (forward and backward)

- Refresh functionality to update timetable/connections

- Arrival time filtering (default is departure) with -a flag

- Help command: SBBuddy -h

## Good to know
- Uppercase letters in terminal shortcut commands are used for main menu options
- Lowercase letters in terminal shortcut commands are used for specific options


## Go Commands

### Run
`` go run .\cmd\SBBuddy\``

### Test
`` go test ./...``

#### Test API
`` go test .\internal\api\ ``

#### Test CLI
`` go test .\internal\ui\ ``

### Build
`` go build -o .\build\ .\cmd\SBBuddy\``

### Distribution
`` goreleaser release --snapshot --clean``

