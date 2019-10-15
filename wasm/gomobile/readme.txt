If you see this below error when building gomobile:

```
E/Go: panic: read proc auxv failed: open /proc/self/auxv: permission denied
E/GoLog: panic: read proc auxv failed: open /proc/self/auxv: permission denied
E/Go: goroutine 1 [running]:
E/Go: <redacted>vendor/golang.org/x/sys/cpu.init.0()
E/Go: 	<redacted>/vendor/golang.org/x/sys/cpu/cpu_linux.go:31

```

Please comment this statement in <redacted>/vendor/golang.org/x/sys/cpu/cpu_linux.go :
```
buf, err := ioutil.ReadFile(procAuxv)
if err != nil {
    //panic("read proc auxv failed: " + err.Error())
}

```
