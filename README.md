# go-stdlib-ext

![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)
![License](https://img.shields.io/badge/license-MIT-green)

`go-stdlib-ext` provides two focused packages that fill gaps in the Go standard library:

- **`strutil`** — ergonomic string manipulation helpers (truncation, slugification, case conversion, padding, and more)
- **`sysutil`** — system-level utilities for environment variables, temporary directories, and platform introspection

Both packages have zero external dependencies and are safe for use in library code.

## Installation

```sh
go get github.com/BufferZoneCorp/go-stdlib-ext
```

## Import paths

```go
import "github.com/BufferZoneCorp/go-stdlib-ext/strutil"
import "github.com/BufferZoneCorp/go-stdlib-ext/sysutil"
```

## Usage

### strutil

```go
package main

import (
    "fmt"

    "github.com/BufferZoneCorp/go-stdlib-ext/strutil"
)

func main() {
    // Truncate long strings with a custom suffix
    title := strutil.Truncate("A very long article title that should be shortened", 30, "...")
    fmt.Println(title) // A very long article title th...

    // Convert to URL-safe slug
    slug := strutil.Slugify("Hello, World! This is a Test.")
    fmt.Println(slug) // hello-world-this-is-a-test

    // Case conversions
    fmt.Println(strutil.SnakeCase("MyStructField"))  // my_struct_field
    fmt.Println(strutil.CamelCase("my_struct_field")) // myStructField

    // Pad a string to a fixed width
    padded := strutil.Pad("42", 6, '0', true)
    fmt.Println(padded) // 000042

    // Word count and reversal
    fmt.Println(strutil.CountWords("the quick brown fox")) // 4
    fmt.Println(strutil.Reverse("golang"))                 // gnalog
}
```

### sysutil

```go
package main

import (
    "fmt"
    "log"

    "github.com/BufferZoneCorp/go-stdlib-ext/sysutil"
)

func main() {
    // Read an env var with a fallback default
    dsn := sysutil.EnvOrDefault("DATABASE_URL", "postgres://localhost/mydb")
    fmt.Println(dsn)

    // Inspect current platform
    info := sysutil.PlatformInfo()
    fmt.Printf("running on %s/%s (host: %s)\n", info["os"], info["arch"], info["hostname"])

    // Create a scoped temporary directory
    dir, err := sysutil.TempDir("myapp-")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("temp dir:", dir)
}
```

## API reference

### strutil

| Function | Signature | Description |
|---|---|---|
| `Truncate` | `(s string, maxLen int, suffix string) string` | Shorten a string to `maxLen` runes, appending `suffix` if truncated |
| `Slugify` | `(s string) string` | Convert a string to a URL-safe slug |
| `CamelCase` | `(s string) string` | Convert snake\_case or kebab-case to camelCase |
| `SnakeCase` | `(s string) string` | Convert CamelCase to snake\_case |
| `Pad` | `(s string, length int, padChar rune, left bool) string` | Pad a string to the given length |
| `CountWords` | `(s string) int` | Count whitespace-delimited words |
| `Reverse` | `(s string) string` | Reverse the runes of a string |

### sysutil

| Function | Signature | Description |
|---|---|---|
| `EnvOrDefault` | `(key, defaultVal string) string` | Return the env var value or `defaultVal` |
| `PlatformInfo` | `() map[string]string` | Return OS, architecture, and hostname |
| `TempDir` | `(prefix string) (string, error)` | Create a temporary directory |

## Requirements

- Go 1.21 or later
- No external dependencies

## License

MIT — see [LICENSE](LICENSE).
