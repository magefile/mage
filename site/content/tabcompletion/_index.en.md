+++
title = "Tab Completion"
weight = 48
+++

Mage supports tab completion for all of its built-in flags as well as your
project's targets. Pressing **Tab** after typing `mage ` will suggest available
targets in the current directory, so you never have to remember exact names.

## Quick Install

The easiest way to get tab completion is the built-in installer. Run
`mage -install` with the name of your shell:

```
mage -install bash
mage -install zsh
mage -install fish
mage -install powershell   # also accepts "pwsh"
```

The installer will:

1. Write a completion script to `~/.config/mage/` (or the appropriate
   platform-specific config directory).
2. Add a source line to your shell's startup file (`.bashrc`, `.zshrc`,
   PowerShell `$PROFILE`, etc.) so completions load automatically in every new
   session.
3. Print a message telling you to restart your shell (or source the config
   file) to activate completions immediately.

If mage can't update your shell config file for any reason, it will print the
line you need to add manually — so the command is always safe to run.

Running `mage -install` again for the same shell is safe. It replaces the
previous completion block rather than duplicating it.

### Shell-Specific Notes

**Bash** — the installer sources the completion script from your `.bashrc`
(preferred) or `.bash_profile`. It uses the standard `complete` built-in, so no
extra packages are required.

**Zsh** — the installer sources the script from your `.zshrc` (honoring
`$ZDOTDIR` if set). If `compdef` is not yet available when the script loads, it
automatically calls `compinit` first.

**Fish** — completions are placed in
`$XDG_CONFIG_HOME/fish/completions/mage.fish` (defaulting to
`~/.config/fish/completions/`). Fish loads files from this directory
automatically, so no startup-file modification is needed.

**PowerShell** — the installer writes a `.ps1` script and sources it from your
`$PROFILE`. Both PowerShell Core (`pwsh`) and Windows PowerShell are supported.
If the profile file or its parent directory doesn't exist yet, mage creates
them.

## The -autocomplete Flag

Under the hood, all of the completion scripts call:

```
mage -autocomplete
```

This prints a plain list of targets (one per line) for the current directory and
exits. You can run it yourself to see what completions would be offered:

```
$ mage -autocomplete
build
clean
deploy
test
```

### Custom / Advanced Usage

If you use a shell that isn't directly supported, or you want to integrate mage
completions into a custom tool, you can wire up `mage -autocomplete` yourself.
The contract is simple:

* It prints target names separated by newlines to stdout.
* It returns exit code 0 on success.
* It reads magefiles from the **current working directory**, so make sure the
  completion function `cd`s to the project directory (or runs mage there) before
  invoking it.

For example, a minimal POSIX-shell completion function might look like:

```sh
_mage_complete() {
    COMPREPLY=( $(mage -autocomplete 2>/dev/null) )
}
complete -F _mage_complete mage
```

Adapt this pattern for any environment that can execute a command and consume its
line-delimited output.
