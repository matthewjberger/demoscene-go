set windows-shell := ["powershell.exe"]

# Displays the list of available commands
@just:
    just --list

# Runs the named app (default: editor). Example: `just run breakout`.
run $project="editor":
    go run ./cmd/{{project}}

# Builds the named app's desktop binary.
build $project="editor":
    go build ./cmd/{{project}}

# Builds the named app's wasm bundle into site/<project>/ (Windows).
[windows]
build-wasm $project="editor":
    New-Item -ItemType Directory -Force -Path site/{{project}} | Out-Null
    $env:GOOS = "js"; $env:GOARCH = "wasm"; go build -o site/{{project}}/main.wasm ./cmd/{{project}}
    Copy-Item "$((go env GOROOT))/lib/wasm/wasm_exec.js" site/{{project}}/wasm_exec.js

# Builds the named app's wasm bundle into site/<project>/ (Unix).
[unix]
build-wasm $project="editor":
    mkdir -p site/{{project}}
    GOOS=js GOARCH=wasm go build -o site/{{project}}/main.wasm ./cmd/{{project}}
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" site/{{project}}/wasm_exec.js

# Serves site/ on http://localhost:8080
serve:
    go run ./cmd/serve

# Builds the named app's wasm bundle and serves site/.
run-wasm $project="editor": (build-wasm project)
    just serve

# Runs go vet and fails on unformatted files (Windows)
[windows]
check:
    go vet ./...
    $unformatted = (gofmt -l . | Out-String).Trim(); if ($unformatted) { Write-Host $unformatted; exit 1 }

# Runs go vet and fails on unformatted files (Unix)
[unix]
check:
    go vet ./...
    unformatted="$(gofmt -l .)"; if [ -n "$unformatted" ]; then echo "$unformatted"; exit 1; fi

# Formats all Go files
format:
    gofmt -w .

# Runs all tests
test:
    go test ./...

# Runs check + test (use this before pushing)
ci: check test

# Lists all module dependencies with available updates
outdated:
    go list -m -u all

# Shows what `go mod tidy` would change without applying it
tidy-check:
    go mod tidy -diff

# Tidies go.mod / go.sum
tidy:
    go mod tidy

# Runs every read-only check: vet+fmt, tidy diff, outdated, tests
audit: check tidy-check outdated test

# Renders package docs
doc:
    go doc -all ./ecs
    go doc -all ./render
    go doc -all ./app

# Removes any built binaries (Windows)
[windows]
clean:
    Remove-Item -Force -ErrorAction SilentlyContinue indigo.exe
    Remove-Item -Force -ErrorAction SilentlyContinue breakout.exe

# Removes any built binaries (Unix)
[unix]
clean:
    rm -f indigo indigo.exe breakout breakout.exe

# Displays Go tool version
@versions:
    go version
