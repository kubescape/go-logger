# GO Logger

In this package we combined different loggers so they can be compiled with a single interface.
This enables switching between the different loggers without changing your code!

We also added OpenTelemetry (otel) spans and logs using helpers and wrappers.

## Supported loggers
* Pretty printer
* [Zap](go.uber.org/zap) with otel support
* Mock (empty logger)
* Icon printer

## TODO
* log
* glog
* klog

## How to use


#### Basic usage

It is possible to simply call the logger without any initialization, the default logger is the `pretty` logger

```go
package main

import logger "github.com/kubescape/go-logger"

func main(){

    logger.L().Info("This is a nice and colorful logger")
    // output: [info] This is a nice and colorful logger
}
```

##### Environment variables

You can change the default logger initialization by setting the appropriate environment variable:
* `KS_LOGGER_NAME`- Set the logger name. The default is `pretty`
* `KS_LOGGER_LEVEL` - Set the log level. The default is `info`


#### Initialize a logger
```go
package main

import logger "github.com/kubescape/go-logger"

func main() {
    // initialize colored logger
    logger.InitLogger("pretty")
    logger.L().Info("This is the pretty logger")
    // output: [info] This is the pretty logger

    // initialize icon logger
    logger.InitLogger("icon")
    logger.L().Info("This is the icon logger")
    // output: ℹ️ This is the icon logger

    // initialize zap (json) logger
    logger.InitLogger("zap")
    logger.L().Info("This is the zap logger")
    // output: {"level":"info","ts":"2022-06-20T19:11:34-04:00","msg":"This is the zap logger"}

    // initialize a mock logger. The mock logger does not print anything
    logger.InitLogger("mock")
    logger.L().Info("This message will not be printed")
    // output:
}
```


#### Adding other information to the log

It is possible to add additional information to the log so as strings, integers, errors, date

```go
package main

import "github.com/kubescape/go-logger/helpers"
import logger "github.com/kubescape/go-logger"

func main(){

    logger.L().Info("ID", helpers.String("name", "my name"), helpers.Int("age", 45), helpers.Interface("address", "address object"))
    // output: [info] ID. name: my name; age: 45; address: address object

}
```


#### Using otel

Once you add this code you can start adding spans and use the zap logger to send events attached to spans.
* spans can be created as [manual instrumentation](https://opentelemetry.io/docs/instrumentation/go/manual/)
* or with [instrumentation plugins](https://uptrace.dev/opentelemetry/instrumentations/?lang=go)
* logs should be attached to a context which contains a span using `.Ctx(ctx)`
* only logs with severity > Warn will send events
* the variable `OTEL_COLLECTOR_SVC` configures where to send otel data with the gRPC protocol
* you can specify `ACCOUNT_ID` to enrich data with it

```go
package main

import (
    logger "github.com/kubescape/go-logger"
    "go.opentelemetry.io/otel"
)

func main() {
    // configure otel
    ctx := logger.InitOtel(logger.L(), "<service>", "<version>")
    defer logger.ShutdownOtel(ctx)

    // create a span
    ctx, span := otel.Tracer("").Start(ctx, "<name of the span>")
    defer span.End()

    if err := cmd.Execute(ctx); err != nil {
        // attach log to the span
        logger.L().Ctx(ctx).Fatal(err.Error())
    }
}
```