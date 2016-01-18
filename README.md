# SmallWorld

SmallWorld is a server program acting as a URL shortener through a RESTful API, written in Go

## Install Redis

Go to http://redis.io/download and follow the procedure that suits your system

## Install Go

Go to https://golang.org/doc/install and follow the procedure that suits your system

Make sure you have created a Go workspace (a dedicated folder for all your work on Go), saved its path as the GOPATH environment variable, as explained in the procedures.

## Download and install SmallWorld

In your Terminal, type :

’cd $GOPATH’
’mkdir src’
’go get github.com/flowg/SmallWorld’
’go install github.com/flowg/SmallWorld’

Now, you should have a bin/ folder at the root of your workspace, at the same level than src/. Inside, you should find the SmallWorld binary.

## Launch it

Go to the bin/ folder and launch the server. In your Terminal, type :

’cd $GOPATH/bin’
’./SmallWorld’

## How to use it ?

This API has 3 endpoints :

# _*Short Link creation*_ :  
# _*Redirection*_ :  
# _*Monitoring*_ :  


