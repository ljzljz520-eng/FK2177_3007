# BUG_REPRO

The following failures were observed while validating the initial project state.
Each section records what failed, how to reproduce it, and the complete command output.
They are preserved intentionally; only failing build gates are omitted from the generated Dockerfile.

## Failure 1: Go test (.)

- Observed problem: `Go test (.)` failed in the initial project state.
- Working directory: `.`
- Command: `cd /app && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -count=1 ./...`
- Exit status: `1`

```text
--- FAIL: TestLoadEmptyArchiveThenAppend (0.00s)
panic: assignment to entry in nil map [recovered, repanicked]

goroutine 12 [running]:
testing.tRunner.func1.2({0x18f0a0, 0x329020})
	/usr/local/go/src/testing/testing.go:1974 +0x1a0
testing.tRunner.func1()
	/usr/local/go/src/testing/testing.go:1977 +0x318
panic({0x18f0a0?, 0x329020?})
	/usr/local/go/src/runtime/panic.go:860 +0x12c
clubscore.(*ActivityList).Append(0x1d148ddd3170, {{0x1be9f0, 0x7}, {0x1d148dd52ec0, 0x10}, {0x1d148dd52ed0, 0xe}, 0x4055000000000000, 0x5e})
	/app/activity.go:77 +0x1b4
clubscore.TestLoadEmptyArchiveThenAppend(0x1d148de96d88)
	/app/activity_test.go:143 +0x170
testing.tRunner(0x1d148de96d88, 0x1ce1c0)
	/usr/local/go/src/testing/testing.go:2036 +0xc4
created by testing.(*T).Run in goroutine 1
	/usr/local/go/src/testing/testing.go:2101 +0x3a8
FAIL	clubscore	0.006s
?   	clubscore/cmd/clubscore	[no test files]
FAIL
```

## Architecture reproduction

### linux/amd64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/clubscore): exit `0`
### linux/arm64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/clubscore): exit `0`
