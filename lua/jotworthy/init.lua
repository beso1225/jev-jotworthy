local M = {}

local defaults = {
  command = "jotworthy",
  command_args = { "--json", "--stdin" },
  threshold = 0.6,
  border = "rounded",
  width = 0.72,
  height = 0.62,
  title = " Jotworthy ",
  today_command = nil,
}

local config = vim.deepcopy(defaults)
local state = {
  bufnr = nil,
  winnr = nil,
  text = nil,
  result = nil,
}

local function notify(message, level)
  vim.notify("jotworthy: " .. message, level or vim.log.levels.INFO)
end

local function close_window()
  if state.winnr and vim.api.nvim_win_is_valid(state.winnr) then
    vim.api.nvim_win_close(state.winnr, true)
  end
  state.winnr = nil
  state.bufnr = nil
end

local function window_size()
  local ui = vim.api.nvim_list_uis()[1]
  if not ui then
    return 1, 1, 0, 0
  end

  local width = math.max(1, math.floor(ui.width * config.width))
  local height = math.max(1, math.floor(ui.height * config.height))
  local row = math.max(0, math.floor((ui.height - height) / 2))
  local col = math.max(0, math.floor((ui.width - width) / 2))
  return width, height, row, col
end

local function open_window(lines, opts)
  close_window()

  local bufnr = vim.api.nvim_create_buf(false, true)
  vim.api.nvim_buf_set_option(bufnr, "bufhidden", "wipe")
  vim.api.nvim_buf_set_lines(bufnr, 0, -1, false, lines or {})

  local width, height, row, col = window_size()
  local winnr = vim.api.nvim_open_win(bufnr, true, {
    relative = "editor",
    width = width,
    height = height,
    row = row,
    col = col,
    style = "minimal",
    border = config.border,
    title = opts and opts.title or config.title,
    title_pos = "center",
  })

  vim.api.nvim_win_set_option(winnr, "wrap", true)
  vim.api.nvim_win_set_option(winnr, "cursorline", true)

  state.bufnr = bufnr
  state.winnr = winnr
  return bufnr, winnr
end

local function set_lines(lines, modifiable)
  if not state.bufnr or not vim.api.nvim_buf_is_valid(state.bufnr) then
    return
  end
  vim.api.nvim_buf_set_option(state.bufnr, "modifiable", true)
  vim.api.nvim_buf_set_lines(state.bufnr, 0, -1, false, lines)
  vim.api.nvim_buf_set_option(state.bufnr, "modifiable", modifiable)
end

local function map(mode, lhs, rhs)
  vim.keymap.set(mode, lhs, rhs, {
    buffer = state.bufnr,
    silent = true,
    nowait = true,
  })
end

local function input_text()
  if not state.bufnr or not vim.api.nvim_buf_is_valid(state.bufnr) then
    return ""
  end
  return table.concat(vim.api.nvim_buf_get_lines(state.bufnr, 0, -1, false), "\n")
end

local function result_error(message)
  set_lines({
    "Jotworthy error",
    "",
    message,
    "",
    "[q] Close",
  }, false)
  map({ "n", "i" }, "q", M.close)
  map({ "n", "i" }, "<Esc>", M.close)
end

local function show_result(result)
  state.result = result
  set_lines(M.format_result(result), false)

  map("n", "q", M.close)
  map("n", "<Esc>", M.close)
  map("n", "w", M.open_today)
  map("n", "<CR>", M.open_today)
end

function M.build_command(opts)
  local command = { opts.command }
  local has_threshold = false
  for _, arg in ipairs(opts.command_args or {}) do
    table.insert(command, arg)
    if arg == "--threshold" then
      has_threshold = true
    end
  end
  if opts.threshold ~= nil and not has_threshold then
    table.insert(command, "--threshold")
    table.insert(command, string.format("%.17g", opts.threshold))
  end
  return command
end

function M.format_result(result)
  local decision = result.write and "KEEP" or "SKIP"
  local score = result.write_score or 0
  local lines = {
    "Jotworthy result",
    "",
    "Decision: " .. decision,
    string.format("Confidence: %.2f", score),
    "",
  }

  if result.kinds and result.kinds[1] then
    table.insert(lines, string.format("Top kind: %s (%.2f)", result.kinds[1].kind, result.kinds[1].score))
  else
    table.insert(lines, "Top kind: unknown")
  end

  table.insert(lines, "")
  table.insert(lines, "[w] Open today's note")
  table.insert(lines, "[q] Close")
  return lines
end

function M.close()
  close_window()
end

function M.resolve_today_command()
  if config.today_command then
    return config.today_command
  end

  if vim.fn.exists(":ObsidianToday") == 2 then
    return "ObsidianToday"
  end
  if vim.fn.exists(":Obsidian") == 2 then
    return "Obsidian today"
  end
  return nil
end

function M.open_today()
  local command = M.resolve_today_command()
  if not command then
    notify("could not find ObsidianToday or Obsidian today", vim.log.levels.ERROR)
    return
  end

  close_window()
  if type(command) == "function" then
    command()
  elseif type(command) == "table" then
    vim.api.nvim_cmd({ cmd = command[1], args = vim.list_slice(command, 2) }, {})
  else
    vim.cmd(command)
  end
end

function M.submit()
  local text = vim.trim(input_text())
  if text == "" then
    notify("input is empty", vim.log.levels.WARN)
    return
  end

  state.text = text
  set_lines({ "Jotworthy", "", "Judging...", "", "Please wait..." }, false)

  local bufnr = state.bufnr
  local command = M.build_command(config)
  vim.system(command, { stdin = text, text = true }, function(obj)
    vim.schedule(function()
      if state.bufnr ~= bufnr then
        return
      end

      if obj.code ~= 0 then
        local message = vim.trim(obj.stderr or "")
        if message == "" then
          message = "backend exited with code " .. tostring(obj.code)
        end
        result_error(message)
        return
      end

      local ok, result = pcall(vim.json.decode, obj.stdout or "")
      if not ok then
        result_error("backend returned invalid JSON: " .. tostring(result))
        return
      end
      show_result(result)
    end)
  end)
end

function M.open(text)
  local lines = {}
  if text and text ~= "" then
    lines = vim.split(text, "\n", { plain = true })
  end

  open_window(lines, { title = " Jotworthy input " })
  vim.api.nvim_buf_set_option(state.bufnr, "filetype", "jotworthy")

  map({ "n", "i" }, "<C-s>", M.submit)
  map({ "n", "i" }, "<Esc>", M.close)
  map("n", "q", M.close)

  vim.cmd("startinsert")
end

function M.setup(opts)
  opts = opts or {}
  if opts.threshold ~= nil and (type(opts.threshold) ~= "number" or opts.threshold < 0 or opts.threshold > 1) then
    error("jotworthy threshold must be a number between 0 and 1")
  end
  config = vim.tbl_deep_extend("force", vim.deepcopy(defaults), opts)
end

return M
