package lua

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/cjoudrey/gluahttp"
	"github.com/gustapinto/from-to/internal/event"
	lua "github.com/yuin/gopher-lua"
	luajson "layeh.com/gopher-json"
)

type Mapper struct {
	logger *slog.Logger
}

func NewMapper() (*Mapper, error) {
	return &Mapper{
		logger: slog.With("mapper", "Lua"),
	}, nil
}

func (m *Mapper) Map(e event.Event) ([]byte, error) {
	l := lua.NewState()
	defer l.Close()

	l.PreloadModule("http", gluahttp.NewHttpModule(&http.Client{}).Loader)
	l.PreloadModule("json", luajson.Loader)

	if err := l.DoFile(e.Metadata.Lua.FilePath); err != nil {
		return nil, err
	}

	m.logger.Debug("Loaded lua file", "file", e.Metadata.Lua.FilePath)

	eventTable := m.eventToLuaTable(l, e)

	err := l.CallByParam(lua.P{
		Fn:      l.GetGlobal(e.Metadata.Lua.Function).(*lua.LFunction),
		NRet:    1,
		Protect: true,
	}, eventTable)
	if err != nil {
		return nil, err
	}

	m.logger.Debug("Executed lua function", "file", e.Metadata.Lua.FilePath, "function", e.Metadata.Lua.Function)

	result := m.toGoValue(l.Get(-1))
	resultMap, ok := result.(map[string]any)
	if !ok {
		return nil, errors.New("failed to convert mapped value back to Go")
	}

	payload, err := json.Marshal(resultMap)
	if err != nil {
		return nil, err
	}

	m.logger.Debug("Converted lua function return to payload", "file", e.Metadata.Lua.FilePath, "function", e.Metadata.Lua.Function, "payload", string(payload))

	return payload, nil
}

func (m *Mapper) toGoValue(luaValue lua.LValue) any {
	switch value := luaValue.(type) {
	case *lua.LNilType:
		return nil
	case lua.LBool:
		return bool(value)
	case lua.LString:
		return string(value)
	case lua.LNumber:
		return float64(value)
	case *lua.LTable:
		maxn := value.MaxN()
		isMap := maxn == 0

		if isMap {
			data := make(map[string]any)
			fmt.Print(data)
			value.ForEach(func(key, value lua.LValue) {
				keystr := fmt.Sprint(m.toGoValue(key))
				data[keystr] = m.toGoValue(value)
			})
			return data
		} else {
			data := make([]any, 0, maxn)
			for i := 1; i <= maxn; i++ {
				data = append(data, m.toGoValue(value.RawGetInt(i)))
			}
			return data
		}
	default:
		return value
	}
}

func (m *Mapper) toLuaValue(l *lua.LState, value any) lua.LValue {
	switch v := value.(type) {
	case string:
		return lua.LString(v)
	case int:
		return lua.LNumber(float64(v))
	case int64:
		return lua.LNumber(float64(v))
	case float64:
		return lua.LNumber(float64(v))
	case map[string]any:
		table := l.NewTable()
		for key, value := range v {
			table.RawSetString(key, m.toLuaValue(l, value))
		}

		return table
	case []any:
		list := l.NewTable()
		for i, value := range v {
			list.RawSet(lua.LNumber(i), m.toLuaValue(l, value))
		}

		return list
	default:
		return nil
	}
}

func (m *Mapper) eventToLuaTable(l *lua.LState, e event.Event) lua.LValue {
	eventMap := map[string]any{
		"id":    e.ID,
		"ts":    e.Ts,
		"op":    e.Op,
		"table": e.Table,
		"row":   e.Row,
	}

	return m.toLuaValue(l, eventMap)
}
