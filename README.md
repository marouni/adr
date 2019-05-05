[![Build Status](https://travis-ci.com/marouni/adr.svg?branch=master)](https://travis-ci.com/marouni/adr)

# ADR Go
A minimalist command line tool written in Go to work with [Architecture Decision Records](http://thinkrelevance.com/blog/2011/11/15/documenting-architecture-decisions) (ADRs).

Greatly inspired by the [adr-tools](https://github.com/npryce/adr-tools) with all of the added benefits of using the Go instead of Bash.

# Quick start
## Installing adr
Go to the [releases page](https://github.com/marouni/adr/releases) and grab one of the binaries that corresponds to your platform.

Alternatively, if you have a Go developement environment setup you can install it directly using :
```bash
go get github.com/marouni/adr && go install github.com/marouni/adr
```


## Initializing adr
Before creating any new ADR you need to choose a folder that will host your ADRs and use the `init` sub-command to initialize the configuration :

```bash
adr init /home/user/my_adrs
```

## Creating a new ADR

As simple as :
```bash
adr new my awesome proposition
```
this will create a new numbered ADR in your ADR folder :
`xxx-my-new-awesome-proposition.md`.
Next, just open the file in your preferred markdown editor and starting writing your ADR.
