# Go SQL Executor

Execute SQL in a Lua VM. `gsql` runs a `.lua` script on the gopher-lua VM and
exposes the `sql` module, which wraps Go's `database/sql` so scripts can open
connections, run queries, and drive transactions.

## Build

```sh
make              # build bin/gsql for the host platform
make build        # cross-compile windows/linux/darwin amd64 into build/
make clean        # remove cross-compiled binaries
make help         # list available targets
```

## Usage

```sh
gsql [options] <script.lua> [extra args...]
```

- `--debug`, `-d` — enable debug logging (nested-logrus formatter at debug level)
- `--version` — print the build version (stamped from `git describe` via `-ldflags`)
- args after the script path are exposed to the script as the global `args`
  array, 1-indexed and including the script path at `args[1]` (`args[2]` is the
  first argument after the script path)
- the script's directory is appended to Lua `package.path`, so sibling `.lua`
  files can be required without an explicit path

## `sql` module

```lua
local sql = require("sql")
```

| Function / method | Returns |
| --- | --- |
| `sql.new()` | new client userdata |
| `sql.escape(s)` | escaped copy of `s` (escapes `\r \n \\ ' " \032`) |
| `c:connect(driver, dsn)` | `ok, err` — opens the connection pool |
| `c:set_timeout(ms)` | set per-operation timeout |
| `c:set_keepalive(idle_ms, pool_size)` | set conn max lifetime and idle pool size |
| `c:close()` | `ok` — close the pool (idempotent) |
| `c:query(sql, ...)` | `rows, err` — `rows` is an array of row tables; extra args are bind values |
| `c:begin_tx()` | `tx, err` — transaction userdata |
| `tx:exec(sql, ...)` | `result, err` — returns exec result |
| `tx:query(sql, ...)` | `rows, err` — same shape as `c:query` |
| `tx:commit()` | `err` — `nil` on success |
| `tx:rollback()` | `err` — `nil` on success |
| `res:last_insert_id()` | `id, err` — last insert id of an exec result |

Driver name is the first argument to `c:connect`, so any `database/sql` driver
registered in the binary works — `mssql` is built in via
`github.com/microsoft/go-mssqldb`.

## Lua example

```lua
local sql = require("sql")

local c = sql.new()
local ok, err = c:connect("mssql",
    "Server=tcp:127.0.0.1,1433;Database=<database>;User ID=<user>;Password=<password>;Integrated Security=false;")
if ok then
    local tx = c:begin_tx()
    if tx ~= nil then
        _, err = tx:exec("update User set name = 'User Name' where UserProfileId=1")
        if err ~= nil then
            print(err)
            tx:rollback()
        end
        tx:commit()
    end
else
    print(err)
end
c:close()
```

## Useful functions

`dumpTable` prints a nested row table (`c:query` / `tx:query` results) one key per line:

```lua
function dumpTable(table, depth)
    if (depth > 200) then
        print("Error: Depth > 200 in dumpTable()")
        return
    end
    for k, v in pairs(table) do
        if (type(v) == "table") then
            print(string.rep("  ", depth) .. k .. ":")
            dumpTable(v, depth + 1)
        else
            print(string.rep("  ", depth) .. k .. ": ", v)
        end
    end
end
```
