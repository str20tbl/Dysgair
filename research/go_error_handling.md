# Go Error Handling Best Practices

**Date**: 2025-11-17
**Context**: Comprehensive error handling audit and improvements for Dysgair project
**References**: Go Proverbs, Effective Go, Go blog

---

## Go Proverbs on Error Handling

### "Don't just check errors, handle them gracefully" - Rob Pike

Error checking is not just about detecting errors—it's about responding appropriately:
- Log for debugging
- Retry when appropriate
- Provide user-friendly messages
- Clean up resources
- Don't ignore errors silently

### "Errors are values" - Rob Pike

Errors in Go are ordinary values that can be:
- Stored in variables
- Passed as arguments
- Returned from functions
- Inspected and acted upon

This makes error handling explicit and flexible.

### "Make the zero value useful" - Go Proverb

When errors occur:
- Zero values (empty strings, 0, nil) should be safe defaults
- Functions should return meaningful zero values on error
- Example: `transcribe()` returns empty map on error, not nil

---

## Error Handling Patterns by Context

### 1. File/Resource Close Errors

#### Pattern A: CRITICAL - Affects Function Success

**When to use**: Close errors mean data corruption or incomplete writes (ZIP files, database transactions)

```go
func CreateZIPArchive(sourceDir, zipPath string) (err error) {
    zipFile, err := os.Create(zipPath)
    if err != nil {
        return err
    }
    defer func() {
        if closeErr := zipFile.Close(); closeErr != nil && err == nil {
            err = closeErr
        }
    }()

    zipWriter := zip.NewWriter(zipFile)
    defer func() {
        // CRITICAL: zipWriter.Close() writes the ZIP central directory
        // If this fails, the entire ZIP file is corrupted
        if closeErr := zipWriter.Close(); closeErr != nil && err == nil {
            err = closeErr
        }
    }()

    // ... rest of function
}
```

**Why**: Named return + defer func allows checking close errors without changing control flow

**Applied in**: `app/services/latex.go:67` (CreateZIPArchive)

#### Pattern B: SHOULD LOG - Operational Awareness

**When to use**: Close errors indicate problems but don't corrupt user data

```go
func readFile(filename string) ([]byte, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer func() {
        if err := file.Close(); err != nil {
            revel.AppLog.Errorf("Failed to close file %s: %v", filename, err)
        }
    }()

    return io.ReadAll(file)
}
```

**Why**: Helps diagnose file descriptor leaks, permission issues, disk problems

**Applied in**: `app/models/wordList.go:63`

#### Pattern C: ACCEPTABLE - Standard Practice

**When to use**: HTTP response body close after full consumption

```go
resp, err := http.Get(url)
if err != nil {
    return err
}
// Note: Body.Close() error intentionally ignored (standard practice)
// Body is fully consumed before close, errors are rare and non-critical
defer resp.Body.Close()
```

**Why**: Go standard library documentation says this is acceptable after body is read

**Applied in**:
- `app/models/transcribe.go:24`
- `app/services/analytics.go:90`

**Reference**: [Go net/http documentation](https://pkg.go.dev/net/http#Response)

---

### 2. Flush/Writer Errors

#### Pattern: Check Error After Flush

```go
writer := csv.NewWriter(w)
defer func() {
    writer.Flush()
    if err := writer.Error(); err != nil {
        revel.AppLog.Errorf("CSV flush error: %v", err)
    }
}()
```

**Why**:
- `csv.Writer.Flush()` doesn't return error
- Must call `writer.Error()` to check if flush succeeded
- Critical for data export integrity

**Applied in**: `app/controllers/transcriptionReview.go:162`

**TODO**: Consider if this should follow LaTeX pattern (named return + propagate error)

---

### 3. Cleanup Errors

#### Pattern: Best-Effort Cleanup with Logging

**When to use**: Temporary directory cleanup, cache cleanup

```go
defer func() {
    // Best-effort cleanup: log errors but don't fail the request
    // (response has already been sent to user)
    if err := os.RemoveAll(tempDir); err != nil {
        revel.AppLog.Errorf("Failed to clean up temp directory %s: %v", tempDir, err)
    }
}()
```

**Why**:
- Response already sent, can't change return value
- Logs help diagnose disk space issues
- Temp files will be cleaned by OS eventually

**Applied in**: `app/controllers/analytics.go:117`

---

### 4. String Parsing Errors

#### Pattern: Log Warning, Use Zero Value

**When to use**: Parsing user input that has a safe default

```go
var userID int64
if _, err := fmt.Sscanf(filter.UserID, "%d", &userID); err != nil {
    revel.AppLog.Warnf("Failed to parse UserID '%s': %v (treating as 0)", filter.UserID, err)
}
args = append(args, userID) // Uses 0 if parse failed
```

**Why**:
- Parse failure → userID stays 0
- Log warns of invalid input
- Query behavior changes are logged (filters by 0 instead of intended ID)

**Applied in**:
- `app/models/dysgair.go:295` (FilteredEntryQuery)
- `app/models/dysgair.go:368` (GetEntriesForAnalysis)

**Alternative**: Could use `strconv.ParseInt()` for explicit error vs. fmt.Sscanf

---

### 5. Intentionally Ignored Errors

#### Pattern A: Fallback in Error Handler

```go
if err != nil {
    var word Word
    // Intentionally ignore error: Fallback query in error handler
    // If this fails too, word will be zero-value which is handled below
    _ = txn.SelectOne(&word, "SELECT * FROM Word WHERE id = ?", id)
    // ... use word (may be empty) ...
}
```

**Why**: Already in error path, double-failure is acceptable

**Applied in**: `app/models/dysgair.go:58`

#### Pattern B: Non-Critical Test Data

```go
for _, item := range testData {
    // Intentionally ignore error: Test data population (non-critical)
    // Duplicates or constraint violations are acceptable during test setup
    _ = dbm.Insert(item)
}
```

**Why**: Test setup can tolerate failures

**Applied in**: `app/controllers/init.go:281,285`

**Best Practice**: Always add comment explaining WHY error is ignored

---

### 6. Functions That Don't Return Errors

These are fine to call without checking:

```go
buf.WriteString("text")         // bytes.Buffer never fails (documented)
entry.UpdateMetricsForBothModels()  // void function
jobs.Now(job)                   // fire-and-forget async
```

**Why**:
- `bytes.Buffer.WriteString()` - Go docs guarantee it never fails (grows buffer automatically)
- Void functions - no error to check
- Fire-and-forget - by design, errors handled elsewhere

---

## Decision Tree: How to Handle Errors

```
Does function return error?
├─ NO → No error handling needed
│   └─ Examples: buf.WriteString(), void functions
│
└─ YES → Check the error
    │
    ├─ Can you recover?
    │   ├─ YES → Handle gracefully
    │   │   └─ Examples: retry, use default, log warning
    │   │
    │   └─ NO → Propagate error up
    │       └─ return fmt.Errorf("context: %w", err)
    │
    └─ Special cases:
        │
        ├─ defer Close() on file/writer
        │   ├─ Data integrity critical? → Named return + check in defer
        │   └─ Not critical? → Log in defer func
        │
        ├─ defer cleanup (temp files, etc)
        │   └─ Log errors (best-effort)
        │
        ├─ HTTP Body.Close()
        │   └─ Ignore with comment (standard practice)
        │
        └─ Intentional ignore
            └─ ALWAYS add comment explaining WHY
```

---

## Academic References

### Go Language Specification
- Error handling philosophy: explicit over implicit
- Errors as ordinary values
- No exceptions - control flow is clear

### Effective Go (golang.org)
- Section: "Errors"
- Multiple return values for error handling
- Error type design patterns

### Go Blog Posts
- **"Error handling and Go"** (blog.golang.org/error-handling-and-go)
  - Error handling strategies
  - Custom error types
  - Wrapping errors

- **"Errors are values"** (blog.golang.org/errors-are-values)
  - Rob Pike's explanation
  - Error handling as ordinary control flow

### Go Proverbs (go-proverbs.github.io)
- "Don't just check errors, handle them gracefully"
- "Errors are values"
- "Don't panic"

---

## Checklist for Adding Error Handling

When you encounter unchecked errors:

1. **Identify the type**
   - [ ] File/resource close
   - [ ] Network operation
   - [ ] String parsing
   - [ ] Cleanup operation
   - [ ] Other

2. **Assess criticality**
   - [ ] CRITICAL - affects data integrity
   - [ ] IMPORTANT - affects user experience
   - [ ] SHOULD LOG - helps ops/debugging
   - [ ] ACCEPTABLE - standard to ignore

3. **Choose pattern**
   - [ ] Named return + check in defer
   - [ ] Log in defer func
   - [ ] Log warning + continue
   - [ ] Propagate error up
   - [ ] Intentional ignore with comment

4. **Document decision**
   - [ ] Add comment explaining choice
   - [ ] Reference this guide if non-obvious

---

## Summary of Improvements Made (2025-11-17)

| File | Location | Pattern Applied | Criticality |
|------|----------|----------------|-------------|
| `services/latex.go` | Line 67 | Named return + defer check | CRITICAL |
| `controllers/analytics.go` | Line 117 | Best-effort cleanup + log | SHOULD LOG |
| `controllers/transcriptionReview.go` | Line 162 | Flush + check Error() | IMPORTANT |
| `models/wordList.go` | Line 63 | Log close error | SHOULD LOG |
| `models/dysgair.go` | Lines 295, 368 | Log parse warnings | SHOULD CHECK |
| `models/transcribe.go` | Line 24 | Document intentional ignore | ACCEPTABLE |
| `services/analytics.go` | Line 90 | Document intentional ignore | ACCEPTABLE |
| `models/dysgair.go` | Line 58 | Document fallback ignore | INTENTIONAL |
| `controllers/init.go` | Lines 281, 285 | Document test data ignore | INTENTIONAL |

**Total errors handled**: 11 locations across 7 files

---

## Future Considerations

### CSV Export Pattern
- **Current**: Logs flush errors but can't propagate (response already sent)
- **Question**: Should it follow LaTeX pattern (named return + check)?
- **Trade-off**: More complex but catches corruption before download
- **Decision**: To be discussed with team

### Global Error Logging
- Consider centralized error logging
- Structured logging (JSON) for better parsing
- Error tracking service (Sentry, Rollbar)

### Error Wrapping
- Go 1.13+ error wrapping with `%w`
- Maintains error chain for debugging
- Example: `fmt.Errorf("failed to parse user ID: %w", err)`

---

**Last Updated**: 2025-11-17
**Status**: ✅ Complete - All error handling improved and documented
