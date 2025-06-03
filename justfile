# Build a target from cmd/<target>/main.go to build/<target>/<target>
build target:
    #!/usr/bin/env bash
    if [ ! -d "cmd/{{target}}" ]; then
        echo "Error: cmd/{{target}} directory does not exist"
        exit 1
    fi
    if [ ! -f "cmd/{{target}}/main.go" ]; then
        echo "Error: cmd/{{target}}/main.go does not exist"
        exit 1
    fi
    go build -o build/{{target}} ./cmd/{{target}}/main.go
    echo "Built {{target}} -> build/{{target}}"

# Run a target, building it first if the executable doesn't exist
run target *args="":
    #!/usr/bin/env bash
    if [ ! -f "build/{{target}}" ]; then
        echo "Executable build/{{target}} not found, building first..."
        just build {{target}}
    fi
    echo "Running {{target}}..."
    ./build/{{target}} {{args}}

# Run a target directly with go run (development mode)
dev target *args="":
    #!/usr/bin/env bash
    if [ ! -d "cmd/{{target}}" ]; then
        echo "Error: cmd/{{target}} directory does not exist"
        exit 1
    fi
    if [ ! -f "cmd/{{target}}/main.go" ]; then
        echo "Error: cmd/{{target}}/main.go does not exist"
        exit 1
    fi
    echo "Running {{target}} in dev mode..."
    go run ./cmd/{{target}}/main.go {{args}}