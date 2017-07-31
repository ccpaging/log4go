2017-07-31

* Move json config and json socklog to `json` as extension

* Move compat with `log` to compat.go

* Replace FileLogWriter with CacheFileLogWriter

2017-07-25

* Compatibility with `log`. New example_log.go to test it

* Add new format "%x", extra short source. Just file without .go suffix 

2017-07-21

* Create new CacheFileLogWriter

* CacheFileLogWriter flush when messages channel length <= 0

* CacheFileLogWriter use bufio as 3rd cache

  Test machine: intel core2 quad, 3.0G speedstep
  Flush size: 8192

```
        BenchmarkFileLog-4               1000000              1978 ns/op
        BenchmarkFileNotLogged-4        20000000               106 ns/op
        BenchmarkFileUtilLog-4            500000              3407 ns/op
        BenchmarkFileUtilNotLog-4        5000000               241 ns/op
```

  FileLog cost reduces 80% compared with no cache, 
  and 80% compared with 2nd cache.

* CacheFileLogWriter use messages channel as 2nd cache

  Test Machine: intel core2 quad, 3.0G speedstep

```
        BenchmarkFileLog-4                300000              6203 ns/op
        BenchmarkFileNotLogged-4        20000000               118 ns/op
        BenchmarkFileUtilLog-4            300000              6353 ns/op
        BenchmarkFileUtilNotLog-4        5000000               263 ns/op
```

  Before

```
        BenchmarkFileLog-4                200000             10692 ns/op
        BenchmarkFileNotLogged-4        20000000               107 ns/op
        BenchmarkFileUtilLog-4            200000              9910 ns/op
        BenchmarkFileUtilNotLog-4        5000000               246 ns/op
```

  FileNotLogged costs more 10%. It may be used on channel transfer.
  FileLog cost reduces -40%.

* move color term log as a extension since it uses 3rd package

* move xmlconfig and xml log as a extension since it uses xml package

2017-07-12

* Fix bug: Initial FileLogWriter.maxbackup = 999

* Restore function parameter: NewFileLogWriter(fname string, rotate bool)

* Campatable https://golang.org/pkg/log function like log.Print(), log.Println() etc. 

2017-05-23

* Change const DefaultFileDepth as var DefaultCallerSkip

2016-03-03

* start goroutine to delete expired log files. Merge from <https://github.com/yougg/log4go>

2016-02-17

* Append log record to current filelog if not oversized

* Fixed Bug: filelog's rename

2015-12-08

* Add maxbackup to filelog

2015-06-09

* Sleeping at most one second and let go routine running drain the log channel before closing

2015-06-01

* Migrate log variables (rec, closeq, closing, etc.) into Filters

* Add new method for Filter include NewFilter(), Close(), run(), WriteToChan()

* When closing, Filter:
  
  + Drain all left log records
  
  + Write them by LogWriter interface
  
  + Then close interface
  
* Every Filter run a routine to recv rec and call LogWriter to write

* LogWrite can be call directly, see log4go_test.go

* Add new method to Logger include skip(), dispatch()

Some ideas come from <https://github.com/ngmoco/timber>. Thanks.

2015-05-12

* Add termlog format. Merge from <https://github.com/alecthomas/log4go>

2015-04-30

* Add closing and wait group. No ugly sleep code.

2015-01-06

* Support json config

* Fixed Bug: lost record in termlog and filelog

2015-01-05 support console color print

* NewConsoleLogWriter() change to NewConsoleLogWriter(color bool)
