# Build

### V1

```shell
task -d scripts build_v1
```

### V2

```shell
task -d scripts build_v2
```

# Run

### V1

```shell
task -d scripts local_v1
```

### V2

```shell
task -d scripts local_v2
```

To debug V2 in IDE:
1) ``` sudo apt update && sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev pkg-config```
2) ```sudo ln -s /usr/lib/x86_64-linux-gnu/pkgconfig/webkit2gtk-4.1.pc /usr/lib/x86_64-linux-gnu/pkgconfig/webkit2gtk-4.0.pc```
3) Add Environment ```PKG_CONFIG_PATH=/usr/lib/x86_64-linux-gnu/pkgconfig``` 
4) Add Go Tool Arguments ```-tags dev```
5) ```task -d scripts local_v2``` and exit after completing
6) Launch ```main.go``` in debug mode.

# Linters

To run linters, use next command:

```shell
 task -d scripts linters -v
```

# Tests

To run test, use next commands. Coverage info will be
recorded to ```coverage``` folder:

```shell
task -d scripts tests -v
```

To include integration tests, add `integration` flag:

```shell
task -d scripts tests integration=true -v
```

# Benchmarks

To run benchmarks, use next command:

```shell
task -d scripts bench -v
```
