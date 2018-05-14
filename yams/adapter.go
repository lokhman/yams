package yams

const (
	AdapterLua = "lua"
)

var Adapters = map[string]string{
	"application/x-lua": AdapterLua,
}

const MaxScriptSize = 1 << 20

const DefaultScript = `local yams = require("yams")

yams.write("YAMS Route: " .. yams.routeid)`
