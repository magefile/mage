+++
title = "Environment Variables"
weight = 40
+++

## MAGEFILE_VERBOSE

Set to "1" or "true" to turn on verbose mode (like running with -v)

## MAGEFILE_DEBUG 

Set to "1" or "true" to turn on debug mode (like running with -debug)

## MAGEFILE_CACHE

Sets the directory where mage will store binaries compiled from magefiles
(default is $HOME/.magefile)

## MAGEFILE_GOCMD

Sets the binary that mage will use to compile with (default is "go").

## MAGEFILE_IGNOREDEFAULT

If set to 1 or true, will tell the compiled magefile to ignore the default
target and print the list of targets when you run `mage`.