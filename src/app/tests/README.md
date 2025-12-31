# Dysgair Test Suite

Comprehensive testing for the Dysgair Welsh pronunciation learning application.

## Test Coverage

### Model Tests (100% Coverage Goal)

#### 1. **model_metrics.go** - MetricsTest Suite
Tests all metric calculation and error attribution algorithms:
- **LevenshteinDistance**: Character-level edit distance with operations tracking
- **LevenshteinWordDistance**: Word-level edit distance
- **CalculateWER**: Word Error Rate calculation (0-100%)
- **CalculateCER**: Character Error Rate calculation (0-100%)
- **AttributeError**: Error source attribution (CORRECT, ASR_ERROR, USER_ERROR, AMBIGUOUS)
- **CalculateTranscriptionAccuracy**: Human vs ASR transcription similarity
- **UpdateMetrics**: Entry metric calculation integration
- **UpdateMetricsForBothModels**: Whisper and Wav2Vec2 comparison

**Key Test Cases:**
- Empty strings
- Identical strings (case insensitive)
- Complete mismatches
- Partial matches
- Welsh language examples
- Unicode/diacritic support
- Edge cases (0%, 100% error rates)

#### 2. **model_users.go** - UserModelTest Suite
Tests user authentication and CRUD operations:
- **NewUser**: User creation with bcrypt email hashing
- **Validate**: Field validation (username length, password match, email format, terms of use)
- **GetUserByUsername**: User retrieval (found/not found)
- **UpdateUserNames**: Name updates with username uniqueness check
- **UpdateEmail**: Email change with password verification
- **UpdatePassword**: Password change with old password verification
- **NewPassword**: Password reset without old password
- **DeleteUser**: User deletion
- **IncrementProgress**: Progress tracking

**Key Test Cases:**
- Password security (bcrypt hashing, verification)
- Email uniqueness
- Username uniqueness
- Validation errors
- Database transactions

#### 3. **model_entries.go** - EntryModelTest Suite
Tests recording and entry management:
- **GetRecordingCount**: Count recordings per user/word
- **IsWordComplete**: 5-recording completion check
- **Entry.Init**: Entry initialization for new/existing users

**Key Test Cases:**
- Zero recordings
- 4 recordings (incomplete)
- 5 recordings (complete)
- 6+ recordings (over-complete)
- User isolation (different users, same word)
- Word isolation (same user, different words)
- Multi-word scenarios

### Controller Integration Tests

#### 4. **controller_auth.go** - AuthenticationTest Suite
Tests authentication flow and protected routes:
- Homepage, registration, login pages
- User registration with validation
- Login with credentials
- Logout
- Password reset
- Protected route access (redirects when not authenticated)
- Static resource serving
- Session handling

**Test Scenarios:**
- Complete registration form
- Missing required fields
- Password mismatch
- Email mismatch
- Username too short
- Terms not accepted
- Invalid email format

#### 5. **controller_dysgair.go** - DysgairControllerTest Suite
Tests main recording application:
- Dysgair index page
- Audio playback (PlayAudio endpoint)
- Recording upload with validation
- Progress navigation (Increment/Decrement/Reset)
- 5-recording system enforcement

**Test Scenarios:**
- Upload without file
- File too small (< 2KB)
- Valid file size (> 2KB)
- Word completion at 5 recordings
- Navigation restrictions (requires 5 recordings)
- Progress boundaries (min 1, max 2500)
- JSON response format

#### 6. **controller_transcription.go** - TranscriptionReviewTest Suite
Tests transcription review admin interface:
- Transcription listing with filters (user, word, error type, review status, date range)
- Human transcription updates
- Metric recalculation
- Mark as reviewed
- CSV/JSON export
- Recording playback
- Bulk metric recalculation

**Test Scenarios:**
- Filter combinations
- Export with filters
- Path traversal security
- Model comparison (Whisper vs Wav2Vec2)
- Reviewer tracking
- 1000 result limit

#### 7. **controller_analytics.go** - AnalyticsTest Suite
Tests analytics dashboard and Python API integration:
- Analysis execution (model comparison, IRR, error attribution, learning curves, word difficulty, correlations)
- LaTeX export with charts
- ZIP file generation
- Download export files

**Test Scenarios:**
- All users vs specific user
- No data available
- No complete words (< 5 recordings)
- Python API integration
- File security (path traversal prevention)

#### 8. **controller_user_mgmt.go** - UserManagementTest Suite
Tests user management admin functions:
- Admin page (user list)
- Profile page
- New user creation
- User deletion
- Password updates
- Email updates
- Name updates

**Test Scenarios:**
- Duplicate email prevention
- Duplicate username prevention
- Password verification
- Bilingual flash messages (Welsh/English)
- Redirect to referer

## Running Tests

### Interactive Mode (Browser UI)

1. **Start the application in dev mode:**
   ```bash
   revel run app dev
   ```

2. **Access the test runner:**
   ```
   http://localhost:9000/@tests
   ```

3. **Features:**
   - Visual test runner interface
   - Hot-reload on code changes
   - Run all tests or specific suites
   - See results in real-time
   - Click to run individual tests

### Command Line Mode

1. **Run all tests:**
   ```bash
   revel test app dev
   ```

2. **Run specific test suite:**
   ```bash
   revel test app dev MetricsTest
   ```

3. **Run specific test method:**
   ```bash
   revel test app dev MetricsTest.TestCalculateWER_EdgeCases
   ```

4. **Test results:**
   - Output in `test-results/` directory
   - `app.log`: Application logs
   - HTML files per test suite
   - `result.passed` or `result.failed` status file
   - Exit code 0 (success) or non-zero (failure) for CI

### Coverage Analysis

Generate coverage reports:

```bash
go test -cover ./app/models/...
go test -coverprofile=coverage.out ./app/models/...
go tool cover -html=coverage.out
```

## Test Organization

### File Structure
```
tests/
├── README.md                     # This file
├── apptest.go                    # Original basic test
├── helpers.go                    # Test utilities and fixtures
├── model_metrics.go              # Metrics algorithm tests
├── model_users.go                # User model tests
├── model_entries.go              # Entry model tests
├── controller_auth.go            # Authentication tests
├── controller_dysgair.go         # Recording app tests
├── controller_transcription.go   # Transcription review tests
├── controller_analytics.go       # Analytics tests
└── controller_user_mgmt.go       # User management tests
```

### Test Naming Convention

- **Test suites**: PascalCase struct embedding `testing.TestSuite`
- **Test methods**: Must start with "Test" (case-sensitive)
- **Format**: `TestFeature_Scenario` (e.g., `TestLogin_WithCredentials`)

### Helper Functions

The `helpers.go` file provides utilities:

- `CreateTestUser(username, email, password)` - Create test user
- `CreateTestWord(text, english)` - Create test word
- `CreateTestEntry(userID, wordID, text, attempt)` - Create test entry
- `CreateTestEntries(userID, wordID, count)` - Create multiple entries (for 5-recording tests)
- `CleanupTestUser(userID)` - Delete test user and related data
- `CleanupTestWord(wordID)` - Delete test word
- `AssertFloatEquals(actual, expected, tolerance)` - Float comparison
- `AssertStringContains(str, substr)` - Substring check
- `GenerateUniqueEmail(prefix)` - Generate unique test email
- `GenerateUniqueUsername(prefix)` - Generate unique test username

## Best Practices

### 1. Test Independence
- Each test should be independent
- Use `Before()` for setup, `After()` for cleanup
- Don't rely on test execution order

### 2. Database Testing
- Tests check if `controllers.Dbm` is available
- Skip database tests gracefully if DB not available
- Use transactions and rollback for isolation

### 3. Authentication Testing
- Most controller tests will redirect when not authenticated
- Use helper functions to create authenticated sessions
- Test both authenticated and unauthenticated scenarios

### 4. Edge Cases
- Test empty inputs
- Test boundary conditions (0, 1, max values)
- Test invalid inputs
- Test security (path traversal, SQL injection prevention)

### 5. Error Handling
- Verify error messages are bilingual (Welsh/English)
- Check flash messages are set correctly
- Ensure graceful failure

## Configuration

Test runner is enabled in `conf/app.conf`:

```ini
[dev]
module.testrunner = github.com/str20tbl/modules/testrunner
```

Routes are configured in `conf/routes`:

```
module:testrunner
```

## Continuous Integration

For CI pipelines, use command-line mode:

```bash
#!/bin/bash
revel test app dev
EXIT_CODE=$?

if [ $EXIT_CODE -ne 0 ]; then
    echo "Tests failed!"
    exit 1
fi

echo "All tests passed!"
exit 0
```

## Test Statistics

- **Total Test Suites**: 8
- **Model Tests**: 3 suites (metrics, users, entries)
- **Controller Tests**: 5 suites (auth, dysgair, transcription, analytics, user_mgmt)
- **Estimated Test Count**: 200+ individual test methods
- **Coverage Goal**: 100% model coverage, comprehensive controller coverage

## Known Limitations

### Database Tests
- Require active database connection
- Skip gracefully if database unavailable
- Use real database (not in-memory mock)

### External Services
- Transcription API tests may fail if service unavailable
- TTS API tests require external service
- Python analytics API requires Python service running

### File System
- Recording upload tests require `/data/recordings/` directory
- Audio playback tests require existing audio files
- LaTeX export tests require `/data/` directory permissions

## Future Enhancements

1. **Mock External Services**
   - Mock transcription API responses
   - Mock TTS API responses
   - Mock Python analytics service

2. **Test Fixtures**
   - SQL fixtures for consistent test data
   - Sample audio files for upload tests
   - Mock database for faster tests

3. **Performance Tests**
   - Benchmark metric calculations
   - Load testing for upload endpoint
   - Concurrent user scenarios

4. **End-to-End Tests**
   - Complete user workflows
   - Multi-user scenarios
   - Full recording session simulation

## Troubleshooting

### Tests Not Appearing in UI

- Ensure test file names don't use `*_test.go` format (Go ignores these)
- Test methods must start with "Test" (case-sensitive)
- Test suite must embed `testing.TestSuite`

### Database Connection Errors

- Check database is running
- Verify connection in `conf/app.conf`
- Tests skip gracefully if DB unavailable

### Test Failures

- Check application logs in `test-results/app.log`
- Verify all dependencies are installed
- Ensure database has required tables

## Documentation

For more information:
- [Revel Testing Documentation](https://revel.github.io/manual/testing.html)
- [Go Testing Package](https://golang.org/pkg/testing/)
- [Revel Framework](https://revel.github.io/)

## Contributing

When adding new tests:

1. Follow naming conventions
2. Add descriptive comments
3. Test both success and failure paths
4. Update this README with new test suites
5. Ensure tests are independent
6. Use helper functions from `helpers.go`

## License

Part of the Dysgair project. See main project license.
