# Go SQL Executor
Execute SQL in Lua VM.


Lua example

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

## Usefull functions
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