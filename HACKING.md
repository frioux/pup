## New Tests

The main tests are in `pup_test.go`.  To add a new test you can write a selector
(with or without flags) and then generate the golden file with something like:

```bash
$ go test -run Main/49 -update
```

Verify that the created file (`testdata/049.golden` in this case) is correct.
