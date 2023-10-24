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
(default is $HOME/. or %USERPROFILE%\magefile)

## MAGEFILE_GOCMD

Sets the binary that mage will use to compile with (default is "go").

## MAGEFILE_IGNOREDEFAULT

If set to "1" or "true", tells the compiled magefile to ignore the default
target and print the list of targets when you run `mage`.

## MAGEFILE_HASHFAST

If set to "1" or "true", tells mage to use a quick hash of magefiles to
determine whether or not the magefile binary needs to be rebuilt. This results
in faster run times (especially on Windows), but means that mage will fail to
rebuild if a dependency has changed. To force a rebuild when you know or suspect
a dependency has changed, run mage with the -f flag.

## MAGEFILE_ENABLE_COLOR

If set to "1" or "true", tells the compiled magefile to print the list of target
when you run `mage` or `mage -l` in ANSI colors.

The default is false for backwards compatibility.

When the value is true and the detected terminal does support colors
then the list of mage targets will be displayed in ANSI color.

When the value is true but the detected terminal does not support colors,
then the list of mage targets will be displayed in the default colors
(e.g. black and white).

## MAGEFILE_TARGET_COLOR

Sets the target ANSI color name which should be used to colorize mage targets.
Only set this when you also set the `MAGEFILE_ENABLE_COLOR` environment
variable to true and want to override the default target ANSI color (Cyan).

The supported ANSI color names are any of these:

- Black
- Red
- Green
- Yellow
- Blue
- Magenta
- Cyan
- White
- BrightBlack
- BrightRed
- BrightGreen
- BrightYellow
- BrightBlue
- BrightMagenta
- BrightCyan
- BrightWhite

The names are case-insensitive.
