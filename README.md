# Jev-JotWorthy

Jev-JotWorthy checks whether a piece of text is worth capturing in an Obsidian
daily note. The repository contains the Go backend and a small Neovim plugin
that presents the check in a floating window.

## Backend

Set `JEV_API_KEY` and run the CLI with a text argument:

```sh
jotworthy "I learned that ..."
```

The Neovim plugin uses the machine-readable form over stdin. The default
threshold is `0.60`; pass `--threshold` to override it:

```sh
printf '%s' "I learned that ..." | jotworthy --json --stdin
printf '%s' "I learned that ..." | jotworthy --json --stdin --threshold 0.75
```

## Neovim plugin

Add this repository to your Neovim runtime path with your plugin manager, then
configure the executable path if `jotworthy` is not on `$PATH`:

```lua
require("jotworthy").setup({
  command = "/absolute/path/to/jotworthy",
  threshold = 0.75,
  -- Optional. The plugin also detects both command styles automatically:
  -- today_command = "ObsidianToday",
})
```

Use `:Jotworthy` to open an input window. A visual selection or command
arguments can be used as the initial text:

```vim
:'<,'>Jotworthy
:Jotworthy A short thought to check
```

Press `<C-s>` to submit, then press `w` or `<CR>` on the result window to open
today's Obsidian note. Press `<Esc>` or `q` to close a window.

The plugin detects the legacy `:ObsidianToday` command and the current
`:Obsidian today` command. Set `today_command` explicitly when a custom command
or a specific Obsidian.nvim version is required.

## Development

Run the Lua smoke test with Neovim:

```sh
nvim --headless --clean -u NONE --cmd 'set rtp^=.' \
  -c 'luafile tests/jotworthy_spec.lua' -c 'qa!'
```

Run the Go tests from the Nix development shell:

```sh
nix develop .#default --command go test ./...
```
