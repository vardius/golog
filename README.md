📟 golog
================
[![Build Status](https://travis-ci.org/vardius/golog.svg?branch=master)](https://travis-ci.org/vardius/golog)
[![Go Report Card](https://goreportcard.com/badge/github.com/vardius/golog)](https://goreportcard.com/report/github.com/vardius/golog)
[![codecov](https://codecov.io/gh/vardius/golog/branch/master/graph/badge.svg)](https://codecov.io/gh/vardius/golog)
[![](https://godoc.org/github.com/vardius/golog?status.svg)](https://pkg.go.dev/github.com/vardius/golog)
[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/vardius/golog/blob/master/LICENSE.md)

<img align="right" height="180px" src="https://github.com/vardius/gorouter/blob/master/website/src/static/img/logo.png?raw=true" alt="logo" />

golog - Logger

📖 ABOUT
==================================================
Contributors:

* [Rafał Lorenz](http://rafallorenz.com)

Want to contribute ? Feel free to send pull requests!

Have problems, bugs, feature ideas?
We are using the github [issue tracker](https://github.com/vardius/golog/issues) to manage them.

## 📚 Documentation

For __examples__ **visit [godoc#pkg-examples](http://godoc.org/github.com/vardius/golog#pkg-examples)**

For **GoDoc** reference, **visit [pkg.go.dev](https://pkg.go.dev/github.com/vardius/golog)**

🚏 HOW TO USE
==================================================

## 🏫 Basic example
```go
package main

import (
    "fmt"
    "context"
    
    "github.com/vardius/golog"
)

func main() {
    ctx := context.Background()
    logger := golog.New()
    
    logger.Debug(ctx context.Context, fmt.Sprintf("Hello %s!", "you"))
}
```

## 📦 As a package
```go
package mylogger

import (
    "context"
    
    "github.com/vardius/golog"
)

var Logger golog.Logger

func SetFlags(flag int) {
    Logger.SetFlags(flag)
}

func SetVerbosity(verbosity golog.Verbose) {
    Logger.SetVerbosity(verbosity)
}

func Debug(ctx context.Context, v string) {
    Logger.Debug(ctx, v)
}

func Info(ctx context.Context, v string) {
    Logger.Info(ctx, v)
}

func Warning(ctx context.Context, v string) {
    Logger.Warning(ctx, v)
}

func Error(ctx context.Context, v string) {
    Logger.Error(ctx, v)
}

func Critical(ctx context.Context, v string) {
	Logger.Critical(ctx, v)
}

func Fatal(ctx context.Context, v string) {
	Logger.Fatal(ctx, v)
}

func init() {
    Logger = golog.New()
}
```
usage:
```go
package main

import (
    "fmt"
    
    "mylogger"
)

func main() {
    mylogger.Debug(ctx context.Context, fmt.Sprintf("Hello %s!", "you"))
}
```

📜 [License](LICENSE.md)
-------

This package is released under the MIT license. See the complete license in the package
