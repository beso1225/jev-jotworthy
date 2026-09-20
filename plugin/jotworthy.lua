if vim.g.loaded_jotworthy then
  return
end
vim.g.loaded_jotworthy = 1

vim.api.nvim_create_user_command("Jotworthy", function(opts)
  local text
  if opts.range > 0 then
    text = table.concat(vim.api.nvim_buf_get_lines(0, opts.line1 - 1, opts.line2, false), "\n")
  elseif opts.args ~= "" then
    text = opts.args
  end
  require("jotworthy").open(text)
end, {
  nargs = "*",
  range = true,
  desc = "Ask Jev whether text belongs in today's Obsidian note",
})
