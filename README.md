# Bazel proxy

A simple proxy for bazel event capture

## Usage

To run the proxy:
``` 
bazel run //cmd/proxy:proxy
```

To run the mapper:
```
bazel run //cmd/mapper:mapper -- -file <filename>
```
