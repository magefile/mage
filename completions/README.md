# mage shell completions

## Zsh

Generate autocomplete script and set it to your `$fpath`:

```
mage -completion zsh > mage.zsh
```

#### Faster autocomplete

The autocompletion script uses `mage -l -format` command for target definitions and descriptions. The built binary is cached but hitting tab
still has a bit of delay for each refresh. To speed up things, you can enable hash fast feature by adding following to your `.zshrc` file:

```
zstyle ':completions:*:mage:*' hash-fast true
```
