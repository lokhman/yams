package yams

const (
	AdapterLua = "lua"
)

var Adapters = map[string]string{
	"application/x-lua": AdapterLua,
}

const DefaultScript = `local yams = require("yams")

yams.write("YAMS Route: " .. yams.routeid)`
