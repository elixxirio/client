# elixxir/client

[![pipeline status](https://gitlab.com/elixxir/client/badges/master/pipeline.svg)](https://gitlab.com/elixxir/client/commits/master)
[![coverage report](https://gitlab.com/elixxir/client/badges/master/coverage.svg)](https://gitlab.com/elixxir/client/commits/master)

This repo contains the Elixxir command-line client (used for integration
testing) and related libraries that facilitate making more full-featured
clients for all platforms.

##Running the Command Line Client

First, make sure dependencies are installed into the vendor folder by running
`glide up`. Then, in the project directory, run `go run main.go`.

If what you're working on requires you to change other repos, you can remove
the other repo from the vendor folder and Go's build tools will look for those
packages in your Go path instead. Knowing which dependencies to remove can be
really helpful if you're changing a lot of repos at once.

If glide isn't working and you don't know why, try removing glide.lock and
~/.glide to brutally cleanse the cache.

Required args:

|Long flag|Short flag|Effect|Example|
|---|---|---|---|
|--numnodes|-n|Number of nodes in each team in the network|-n 3|
|--userid|-i|ID of the user of this client|-i 5|

Optional args:

|Long flag|Short flag|Effect|Example|
|---|---|---|---|
|--gwaddr|-g|Address of the gateway to connect to (Overrides config file)|-g localhost:8443|
|--destid|-d|ID of the user to send messages to|-d 6|
|--message|-m|Text message to send|-m "let's both have a good day"|
|--verbose|-v|Prints more logging messages for debugging|-v|
|--version|-V|Show the generated version information. Run `$ go generate cmd/version.go` if the information is out of date.|--version|
|--sessionfile|-f|File path for storing the session. If not specified, the session will be stored in RAM and won't persist.|-f mySuperCoolSessionFile|
|--noBlockingTransmission| |Disables transmission rate limiting (useful for dummy client)|--noBlockingTransmission|
|--mint| |Creates some coins for this user for testing and demos|--mint|
|--help|-h|Prints a help message with all of these flags|-h|
|--certpath|-c|Enables TLS by passing in path to the gateway certificate file|-c "~/Documents/gateway.cert"|
|--dummyfrequency| |How often dummy messages should be sent per second. This flag is likely to be replaced when we implement better dummy message sending.|--dummyfrequency 0.5|

##Project Structure

`api` package contains functions that clients written in Go should call to do
all of the main interactions with the client library.

`bindings` package exists for compatibility with Gomobile. All functions and
structs in the `bindings` package must be able to be bound with `$ gomobile bind`
or they will be unceremoniously removed. There are many requirements for 
this, and if you're writing bindings, you should check the `gomobile` 
documentation listed below.

In general, clients written in Go should use the `api` package and clients 
written in other languages should use the `bindings` package.

`bots` contains code for interacting with bots. If the amount of code required
to easily interact with a bot is reasonably small, it should go in this package.

`cmd` contains the command line client itself, including the dummy messaging
prototype that sends messages at a constant rate.

`crypto` contains code for encrypting and decrypting individual messages with
the client's part of the cipher. 

`globals` contains a few global variables. Avoid putting more things in here
without seriously considering the alternatives. Most important is the Log 
variable:

```go
globals.Log.ERROR.Println("this is an error")
```

Using this global Log variable allows external users of jww logging, like the 
console UI, to see and print log messages from the client library if they need
to, so please use globals.Log for all logging messages to make this behavior
work consistently.

If you think you can come up with a better design to deal with this problem, 
please go ahead and implement it. Anything that moves towards the globals 
package no longer existing is probably a win.

`io` contains functions for communicating between the client and the gateways.
It's also currently responsible for putting fragmented messages back together.

`parse` contains functions for serializing and deserializing various specialized
information into messages. This includes message types and fragmenting messages
that are too long.

`payment` deals with the wallet and payments, and keeping track of all related
data in non-volatile storage.

`switchboard` includes a structure that you can use to listen to incoming 
messages and dispatch them to the correct handlers.

`user` includes objects that deal with the user's identity and the session 
and session storage.

##Gomobile

We bind all exported symbols from the bindings package for use on mobile 
platforms. To set up Gomobile for Android, install the NDK and 
pass the -ndk flag to ` $ gomobile init`. Other repositories that use Gomobile
for binding should include a shell script that creates the bindings.

###Recommended Reading for Gomobile

https://godoc.org/golang.org/x/mobile/cmd/gomobile (setup and available 
subcommands)

https://godoc.org/golang.org/x/mobile/cmd/gobind (reference cycles, type 
restrictions)

Currently we aren't using reverse bindings, i.e. calling mobile from Go.

###Testing Bindings via Gomobile

The separate `bindings-integration` repository exists to make it easier to 
automatically test bindings. Writing instrumented tests from Android allows 
you to create black-box tests that also prove that all the methods you think 
are getting bound are indeed bound, rather than silently getting skipped.

You can also verify that all symbols got bound by unzipping `bindings-sources.jar`
and inspecting the resulting source files.

Every time you make a change to the client or bindings, you must rebuild the 
client bindings into a .aar to propagate those changes to the app. There's a 
script that runs gomobile for you in the `bindings-integration` repository.
