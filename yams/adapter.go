package yams

const (
	AdapterLua = "lua"
)

type adapterMap map[string]string

func (am adapterMap) GetMimeType(adapter string) string {
	for mimeType, a := range am {
		if a == adapter {
			return mimeType
		}
	}
	return ""
}

var Adapters = adapterMap{
	"application/x-lua": AdapterLua,
}

const DefaultScript = `local yams = require("yams")

yams.write("YAMS Route: " .. yams.routeid)
`
