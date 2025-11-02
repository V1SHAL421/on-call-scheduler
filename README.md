# On-Call Scheduler

A Go application that generates on-call schedules by combining a base rotation plan with override intervals.

## Features

- **Rotating Schedule Generation**: Creates schedules based on user rotation and handover intervals
- **Override Support**: Applies temporary schedule changes that override the base rotation
- **Time Window Filtering**: Trims output to specified date ranges
- **Zero-Duration Filtering**: Automatically removes invalid intervals
- **JSON Input/Output**: Structured data format for easy integration

## Example Usage

```bash
./render-schedule --schedule=schedule.json --overrides=overrides.json --from=2025-11-07T17:00:00Z --until=2025-11-21T17:00:00Z
```

### Command Line Arguments

- `--schedule`: Path to JSON file containing the base schedule plan
- `--overrides`: Path to JSON file containing override intervals
- `--from`: Start time for the output schedule (RFC3339 format)
- `--until`: End time for the output schedule (RFC3339 format)

## Example Input Format

### Schedule Plan (`schedule.json`)
```json
{
  "handover_interval_days": 7,
  "users": ["alice", "bob", "charlie"]
}
```

### Overrides (`overrides.json`)
```json
[
  {
    "user": "charlie",
    "from": "2025-11-10T17:00:00Z",
    "to": "2025-11-10T22:00:00Z"
  }
]
```

## Example Output Format

The application outputs structured JSON logs including the final schedule:

```json
{
  "time": "2025-11-01T17:53:55.993161Z",
  "level": "INFO",
  "msg": "Final schedule",
  "finalSchedule": [
    {
      "user": "alice",
      "start_at": "2025-11-07T17:00:00Z",
      "end_at": "2025-11-10T17:00:00Z"
    }
  ]
}
```

## Building

### Using Make (Recommended)
```bash
# Build the application
make build

# Run tests
make test

# Clean build artifacts
make clean

# Run with example data
make run
```

### Manual Build
```bash
cd src && go build -o ../render-schedule *.go
```

## Project Structure

```
on-call-scheduler/
├── src/           # Source code
│   ├── main.go    # Main application entry point
│   ├── utils.go   # Core scheduling functions
│   └── types.go   # Data type definitions
├── test/          # Test files
│   └── utils_test.go
├── schedule.json  # Example schedule configuration
├── overrides.json # Example override configuration
├── go.mod         # Go module definition
├── Makefile       # Build automation
└── README.md      # This file
```

## Testing

Run the test suite:
```bash
make test
# or
go test ./test/...
```

## Algorithm

1. **Initial Schedule Generation**: Creates a rotating schedule based on the handover interval
2. **Override Application**: Applies overrides by splitting and replacing schedule intervals
3. **Time Window Filtering**: Trims the schedule to the requested time range
4. **Zero-Duration Cleanup**: Removes any intervals with identical start/end times

## Development

### Dependencies
- Go 1.21+
- testify (for testing)

### Adding Tests
Tests are located in the `test/` directory and use the testify framework. All core functions are exported from the `src` package for testing.

## Error Handling

The application validates:
- All required command line arguments
- JSON file format and structure
- Time format (RFC3339)
- Schedule configuration (positive handover intervals, non-empty user list)
- Time window validity (from < until)