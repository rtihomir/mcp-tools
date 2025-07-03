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

# Build a static standalone binary (no external C library dependencies)
build-static target:
    #!/usr/bin/env bash
    if [ ! -d "cmd/{{target}}" ]; then
        echo "Error: cmd/{{target}} directory does not exist"
        exit 1
    fi
    if [ ! -f "cmd/{{target}}/main.go" ]; then
        echo "Error: cmd/{{target}}/main.go does not exist"
        exit 1
    fi
    echo "Building static binary for {{target}}..."
    
    # Check if we're on Linux (supports fully static binaries)
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo "Building fully static binary for Linux..."
        CGO_ENABLED=1 go build \
            -a \
            -ldflags '-extldflags "-static"' \
            -tags netgo,osusergo \
            -o build/{{target}}-static \
            ./cmd/{{target}}/main.go
    else
        # macOS/Windows: build with embedded C libs but not fully static
        echo "Building semi-static binary for {{target}} (embedded C libs)..."
        CGO_ENABLED=1 go build \
            -a \
            -ldflags '-s -w' \
            -tags netgo,osusergo \
            -o build/{{target}}-static \
            ./cmd/{{target}}/main.go
    fi
    echo "Built static {{target}} -> build/{{target}}-static"
    
# Build static binary for distribution (optimized, stripped)
build-release target:
    #!/usr/bin/env bash
    if [ ! -d "cmd/{{target}}" ]; then
        echo "Error: cmd/{{target}} directory does not exist"
        exit 1
    fi
    if [ ! -f "cmd/{{target}}/main.go" ]; then
        echo "Error: cmd/{{target}}/main.go does not exist"
        exit 1
    fi
    echo "Building release binary for {{target}}..."
    
    # Check if we're on Linux (supports fully static binaries)
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo "Building fully static release binary for Linux..."
        CGO_ENABLED=1 go build \
            -a \
            -ldflags '-extldflags "-static" -s -w' \
            -tags netgo,osusergo \
            -trimpath \
            -o build/{{target}}-release \
            ./cmd/{{target}}/main.go
    else
        # macOS/Windows: optimized but not fully static
        echo "Building optimized release binary for {{target}}..."
        CGO_ENABLED=1 go build \
            -a \
            -ldflags '-s -w' \
            -tags netgo,osusergo \
            -trimpath \
            -o build/{{target}}-release \
            ./cmd/{{target}}/main.go
    fi
    echo "Built release {{target}} -> build/{{target}}-release"
    echo "Binary size: $(du -h build/{{target}}-release | cut -f1)"

# Cross-compile for Linux with fully static linking (requires Docker)
build-linux-static target:
    #!/usr/bin/env bash
    if [ ! -d "cmd/{{target}}" ]; then
        echo "Error: cmd/{{target}} directory does not exist"
        exit 1
    fi
    if [ ! -f "cmd/{{target}}/main.go" ]; then
        echo "Error: cmd/{{target}}/main.go does not exist"
        exit 1
    fi
    echo "Cross-compiling fully static binary for Linux..."
    docker run --rm -v $(pwd):/workspace -w /workspace \
        golang:1.24-alpine sh -c '
        apk add --no-cache gcc musl-dev
        CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
            -a \
            -ldflags "-extldflags \"-static\" -s -w" \
            -tags netgo,osusergo \
            -trimpath \
            -o build/{{target}}-linux-static \
            ./cmd/{{target}}/main.go
    '
    echo "Built Linux static {{target}} -> build/{{target}}-linux-static"
    echo "Binary size: $(du -h build/{{target}}-linux-static | cut -f1)"

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