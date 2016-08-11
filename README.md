# polidor

[![Build Status](https://travis-ci.org/soniah/polidor.svg?branch=master)](https://travis-ci.org/soniah/polidor)
[![GoDoc](https://godoc.org/github.com/soniah/polidor?status.png)](http://godoc.org/github.com/soniah/polidor)

Polidor is a program and library for cleaning up a disk based storage hierachy.

Sonia Hamilton, sonia@snowfrog.net, http://www.snowfrog.net.

## Overview

Polidor is Portuguese for "polisher".

The files in Polidor are:

* `cleaner/cleaner.go` and `cleaner/retentions.yml` - the main program
  and it's configuration file. Directories being scanned are printed to
  the screen, to give an indication of progress. Run using:

    % cd cleaner
    % go run cleaner.go

* `generator/generator.go` and `generator/generator.yml` - for
  generating test data for the cleaner program, and it's configuration
  file. The names of files being generated are printed to the screen, to
  give an indication of progress. Run using:

    % cd generator
    % go run generator.go

* `polidor.go` and `polidor_test.go` - library and tests

## Program Logic

* the configuration file is read in
* the storage directory is walked using filepath.Walk()
* to prevent the program taking too long, a channel named `timeout` is used
* the `polidor` library provides various directory parsing functions

## Tests

Tests are run using Travis https://travis-ci.org/soniah/polidor. The can
also be run by:

    % go test
    % go test -run <test name>
