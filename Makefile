.PHONY: build test clean run

build:
	go build -o render-schedule *.go

test:
	go test ./test/...

clean:
	rm -f render-schedule

run: build
	./render-schedule --schedule=schedule.json --overrides=overrides.json --from='2025-11-07T17:00:00Z' --until='2025-11-21T17:00:00Z'

deps:
	go mod tidy

zip:
	git archive --format=zip --output=on-call-scheduler.zip HEAD