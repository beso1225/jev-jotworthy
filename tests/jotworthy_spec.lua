local jotworthy = require("jotworthy")

local function assert_equal(got, want, message)
  if got ~= want then
    error(string.format("%s: got %q, want %q", message, got, want))
  end
end

local command = jotworthy.build_command({
  command = "jotworthy",
  command_args = { "--json", "--stdin" },
})
assert_equal(command[1], "jotworthy", "command executable")
assert_equal(command[2], "--json", "first command argument")
assert_equal(command[3], "--stdin", "second command argument")

local lines = jotworthy.format_result({
  write = true,
  write_score = 0.87,
  kinds = {
    { kind = "idea", score = 0.61 },
    { kind = "learning", score = 0.22 },
  },
})
assert_equal(lines[3], "Decision: KEEP", "positive decision")
assert_equal(lines[4], "Confidence: 0.87", "confidence")
assert_equal(lines[6], "Top kind: idea (0.61)", "top kind")

local skipped = jotworthy.format_result({
  write = false,
  write_score = 0.21,
  kinds = {},
})
assert_equal(skipped[3], "Decision: SKIP", "negative decision")

print("jotworthy Lua tests passed")
