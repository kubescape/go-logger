# GO Logger

In this package we combined different loggers so they can be compiled with a single interface.
This enables switching between the different loggers without changing your code!

## Supported loggers
* Pretty printer
* [Zap](go.uber.org/zap)
* Mock (empty logger)

## TODO
* log
* glog
* klog

## How to use


#### Basic usage

It is possible to simply call the logger without any initialization, the default logger is the `pretty` logger

```go
package main

import logger "github.com/armosec/go-logger"

func main(){

    logger.L().Info("This is a nice and colorful logger")
	// output: [info] This is a nice and colorful logger
}
```

#### Initialize a logger
```go
package main

import logger "github.com/armosec/go-logger"

func main() {
	// initialize colored logger
	logger.InitLogger("pretty")
	logger.L().Info("This is the pretty logger")
	// output: [info] This is the pretty logger

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

import "github.com/armosec/go-logger/helpers"
import logger "github.com/armosec/go-logger"

func main(){

    logger.L().Info("ID", helpers.String("name", "my name"), helpers.Int("age", 45), helpers.Interface("address", "address object"))
    // output: [info] ID. name: my name; age: 45; address: address object

}
```