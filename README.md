# SmallWorld

SmallWorld is a server program acting as a URL shortener through a RESTful API, written in Go

## Install Redis

Go to http://redis.io/download and follow the procedure that suits your system

## Install Go

Go to https://golang.org/doc/install and follow the procedure that suits your system

Make sure you have created a Go workspace (a dedicated folder for all your work on Go), saved its path as the GOPATH environment variable, as explained in the procedures.

## Download and install SmallWorld

In your Terminal, type :

```
    cd $GOPATH
```
```
    mkdir src
```
```
    go get github.com/flowg/SmallWorld
```
```
    go install github.com/flowg/SmallWorld
```

Now, you should have a bin/ folder at the root of your workspace, at the same level than src/. Inside, you should find the SmallWorld binary.

## Launch it

Go to the bin/ folder and launch the server. In your Terminal, type :

```
    cd $GOPATH/bin
```
```
    ./SmallWorld
```

## How to use it ?

This API has 3 endpoints :

+ _*Short Link creation*_ :  send a _POST_ request to *http://your_domain.com/shortlink* with a payload like `url=http://firstcodeafter.com/` to get back a JSON with a token key. The value of this key is your shortlink. A shortlink is only valid for 3 months, it will be deleted afterwards. You cal also send your _POST_ request with this payload `url=http://firstcodeafter.com/&custom=greek` and `greek` will be used to create a customized shortlink.
+ _*Redirection*_ :  type *http://your_domain.com/value_of_the_previous_token* in your browser and you'll be redirected to the URL given in the previous payload. A count of the redirections is stored in the hash representing this shortlink.
+ _*Monitoring*_ :  send a _GET_ request to *http://your_domain.com/admin/value_of_the_previous_token* to get back the hash representing this shortlink as a JSON. That way, you'll be able to see its number of redirections.

You can also take a look at smallWorld.log (all errors and other infos will be logged there) and config.yml (simple configuration file). They will be located in the same folder than the source code file SmallWorld.go .

By default, the number of characters used to generate the shortlink is 6, you can modify it in the config file, as well as the listening port for the server and the datastore configuration. As the Go package Viper (https://github.com/spf13/viper) is used to watch the config file, you shouldn't have to re-start the server to see changes come alive.


