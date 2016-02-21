# COMP8005.ScaleableServer
This project was written in Go and is intended to aid in comparing the preformance of epoll, select and a traditional multi-threaded network archetecture using go routines. This repository contains the default multithreaded scalable server. The other two repositories in this collection can be found here:
* [epoll](https://github.com/mvouve/COMP8005.EPollScalableServer)
* [select](https://github.com/mvouve/COMP8005.SelectScalableServer)

I have also written a [client](https://github.com/mvouve/COMP8005.ScalableServerClient) that works with all three servers. 

##Usage
This server can be envoked using the syntax of:
```bash
./COMP8005.ScalableServer [:Port]
```

When terminated, the process will exit and generate an XLSX report listing clients that had connected, the ammount of data that they transfered and the number of times they transfered data to the server as well as other useful information about the connections.

##Testing
This program has been tested to work on Fedora 22 and Manjaro 15 using a standered Go 1.5 compiler. It has been able to sustain over 40k concurrent connections.
