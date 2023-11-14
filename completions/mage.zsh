#compdef mage

local curcontext="$curcontext" state line ret=1
typeset -A opt_args
local magepath="magefiles"
local -a targets

_arguments -C \
  '-clean[clean out old generated binaries from CACHE_DIR]' \
  '-compile[output a static binary to the given path]:compilepath:_path_files' \
  '-completion[print out shell completion script for given shell (default: "")]' \
  '-init[create a starting template if no mage files exist]' \
  '-h[show help]' \
  '-l[list mage targets in this directory]' \
  '-format[when used in conjunction with -l, will list targets in a specified golang template format (available vars: .Name, .Synopsis)]' \
  '-d[directory to read magefiles from (default "." or "magefiles" if exists)]:magepath:_path_files -/' \
  '-debug[turn on debug messages]' \
  '-f[force recreation of compiled magefile]' \
  '-goarch[sets the GOARCH for the binary created by -compile (default: current arch)]' \
  '-gocmd[use the given go binary to compile the output (default: "go")]' \
  '-goos[sets the GOOS for the binary created by -compile (default: current OS)]' \
  '-ldflags[sets the ldflags for the binary created by -compile (default: "")]' \
  '-h[show description of a target]' \
  '-keep[keep intermediate mage files around after running]' \
  '-t[timeout in duration parsable format (e.g. 5m30s)]' \
  '-v[show verbose output when running mage targets]' \
  '-w[working directory where magefiles will run (default -d value)]' \
  '*: :->trg'

(( $+opt_args[-d] )) && magepath=$opt_args[-d]

zstyle ':completion:*:mage:*' list-grouped false
zstyle -s ':completion:*:mage:*' hash-fast hash_fast false

_get_targets() {
  # check if magefile exists
  [[ ! -f "$magepath/mage.go" ]] && return 1

  local IFS=$'\n'
  targets=($(MAGEFILE_HASHFAST=$hash_fast mage -d $magepath -l -format "{{ .Name }}|{{ .Synopsis }}" | awk -F '|' '{
    target = $1;
    gsub(/:/, "\\:", target);
    gsub(/^ +| +$/, "", $0);
  
    description = $2;
    gsub(/^ +| +$/, "", description);
  
    print target ":" description;
  }'))
}

case "$state" in
  trg)
    _get_targets || ret=1 
    _describe 'mage' targets && ret=0
    ;;
esac

return ret

# Local Variables:
# mode: Shell-Script
# sh-indentation: 2
# indent-tabs-mode: nil
# sh-basic-offset: 2
# End:
# vim: ft=zsh sw=2 ts=2 et
