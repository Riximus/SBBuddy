# SBBuddy

## Current Functionalities
- Look up timetable of a station
- Search connections of two stations
- Fetch a random connection using `-R`. Station names are selected from `internal/api/major_stations.txt` which lists common stations across Switzerland.
- Timetable via command line: `SBBuddy -T "Basel SBB"`
- Connections via command line: `SBBuddy -C "Basel SBB" -C "Zürich HB"`
- Timetable with date/time: `SBBuddy -T "Basel SBB" -d 2025-07-28 -t 14:50`

## Good to know
- Uppercase letters in terminal shortcut commands are used for main menu options
- Lowercase letters in terminal shortcut commands are used for specific options

## TODOS
- [x] Replace the current showing of commands available with the Bubbles library with their Help component (https://github.com/charmbracelet/bubbles)
- [x] Add "via" functionality so that the user can give a start location, one or multiple via locations where they want to pass through and finally the destination they want to get to.
- [x] Add a "random" functionality so that the user can get a random connection between two stations and possibly if declared how many via connections they want to have.
- [x] Add a "quit" functionality so that the user can quit the program.
- [ ] Add Alarm or Notification functionality on a chosen connection so that the user can get notified when the train is about to leave, the user can decide how many minutes they wish to be notified before the train leaves.
- [x] Add the possibility to look up the timetable or connections from a specific time and date.
- [x] The user should be able to scroll (maybe with left and right action) through the connection list to see more connections from before (the first connection on the list) or after (the last connection on the list) the current list
- [x] Add a "refresh" functionality so that the user can refresh the timetable or connections.
- [x] Make keys consistent that the user doesn't get confused. Using the same keys for all the commands. Don't use "any key to continue"
- [x] Create shortcuts for the commands so the user can use, for example:
  - [x] Timetable: "SBBuddy -T Basel SBB"
  - [x] Connection: "SBBuddy -C Basel SBB -C Zürich HB"
  - [x] Connections with multiple connections that takes them in order and between the first and last one are "via": "SBBuddy -C "Basel SBB" -C "Zürich HB" -C Chur"
  - [x] Timetable at a certain time: "SBBuddy -T Basel SBB -t 10:00"
  - [x] Timetable from a certain date and time: "SBBuddy -T Basel SBB -d 14.08.2025 -t 10:00"
  - [x] Connections from a certain date: "SBBuddy -C "Basel SBB" -C "Zürich HB" -d 14.08.2025"
  - [x] Connections from a certain date and time: "SBBuddy -C "Basel SBB" -C "Zürich HB" -d 14.08.2025 -t 10:00"
  - [x] Random connection: "SBBuddy -R" or "SBBuddy -R 2". Combine with `-C` to fix the start or end station.
  - [x] Add a command so the user can set that they want times for arrival (default is departure) open for suggestions what letter to use maybe only -a without value
  - [x] Help: "SBBuddy -h"
- [x] Add the "Help" command so that the user can get a list of all the commands (see above) and their usage with examples.
- [x] Change the tables to use the bubbles library and still be interactive.

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

