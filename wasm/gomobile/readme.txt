When you build gomobile for Android/iOS, maybe you see these following issues:

1. panic: read proc auxv failed: open /proc/self/auxv: permission denied:

```
E/Go: panic: read proc auxv failed: open /proc/self/auxv: permission denied
E/GoLog: panic: read proc auxv failed: open /proc/self/auxv: permission denied
E/Go: goroutine 1 [running]:
E/Go: <redacted>vendor/golang.org/x/sys/cpu.init.0()
E/Go: 	<redacted>/vendor/golang.org/x/sys/cpu/cpu_linux.go:31

```

Please replace file <redacted>/vendor/golang.org/x/sys/cpu/cpu_linux.go by https://github.com/golang/sys/blob/master/cpu/cpu_linux.go :
```
buf, err := ioutil.ReadFile(procAuxv)
if err != nil {
    //panic("read proc auxv failed: " + err.Error())
}

```

2. `libproc.h` file not found
Please comment 2 files: metadata/issuingethrequest.go and metadata/issuingethresponse.go

