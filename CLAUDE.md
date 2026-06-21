# clank

Long-running service. Module `github.com/ianeff/clank`. Structured `slog` logging, context-driven graceful shutdown.

## Definition of done
- `make ci` is green: fmt-check → vet → lint → test (`-race`) → build. Run checks/tests incrementally during edits.
- `make vulncheck` is clean — a separate security gate (govulncheck over deps), not part of `make ci`.

## Go house rules
- Errors: wrap with `%w`, compare with `errors.Is` / `errors.As`, combine with `errors.Join`. Package-level `var ErrFoo = errors.New(...)` for sentinels.
- Never return a typed-nil pointer as an `error` — return literal `nil`.
- Accept interfaces, return structs. Interfaces are consumer-defined, not shipped with the implementation.
- `context.Context` is the first parameter, never a struct field. Thread it through; no `context.Background()` deep in call chains.
- Run `go test -race` for concurrency. Use `testing/synctest` (`synctest.Test`) for deterministic time/concurrency tests.
- Benchmark with `testing.B` and `benchstat` before/after. Check escape analysis via `go build -gcflags=-m`.
- Use stdlib: `any` (not `interface{}`), builtins (`min`/`max`/`clear`), `log/slog`, `slices`/`maps` over hand-rolled loops.
- Don't guess signatures or find-replace blindly — use `go doc` or gopls/LSP tools (`go_rename_symbol`, etc.).

## Service shape
- Operational output goes through the default `slog` JSON handler — no `fmt.Println`.
- Shutdown is driven by `signal.NotifyContext`; new long-running work selects on `ctx.Done()` and exits cleanly.
