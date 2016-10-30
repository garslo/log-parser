# log-parser

Quick & dirty program to parse log lines which have time information
in a `<key>=<value>` format.

Given a file with log lines which contain key-value pairs like

```
mode=<something>
```

or

```
<key>=<value>
```

and corresponding key-value pairs of any of

```
duration_ms=<millisecond_count>
exec_ms=<millisecond_count>
exec=<duration>
duration=<duration>
time_ms=<millisecond_count>
time=<duration>
```

invoking

```
$ log-parser -key mode value <something> <list of files>
```

will print to stdout all the times in milliseconds, one per line. The
files are read line-by-line, so `log-parser` can handle files of
arbitrary size. Each file gets its own worker, so you can speed up
processing of a single, large file by splitting it into several
smaller files.

As a crude benchmark, processing a `1G` file split into 10 `100MB`
files took around 10 seconds (no precise timing), consumed 800%+ cpu,
and used minimal resident memory (< 10MB).
