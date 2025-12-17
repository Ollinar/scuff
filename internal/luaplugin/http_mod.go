package luaplugin

import (
	"io"
	"net/http"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

func WithHTTPModule(client *http.Client) LuaModule {
	return func(L *lua.LState) {
		L.PreloadModule("http", HTTPModuleLoader(client))
	}
}

func HTTPModuleLoader(client *http.Client) lua.LGFunction {
	return func(L *lua.LState) int {
		httpModFunc := map[string]lua.LGFunction{
			"request": httpSendRequest(client),
		}
		mod := L.SetFuncs(L.NewTable(), httpModFunc)
		L.Push(mod)
		return 1
	}
}

// function signiture would be:
// responseBody,responseCode,ResponseHeader,err = request(method,url,{headers:[]""|"",body:""})
func httpSendRequest(client *http.Client) lua.LGFunction {
	return func(L *lua.LState) int {
		method := L.CheckString(1)
		urlStr := L.CheckString(2)
		lv := L.Get(3)

		var reqBody io.Reader
		var headers map[string][]string

		tbl, ok := lv.(*lua.LTable)
		if ok {
			headers = httpHeadersFromLua(tbl)

			bodyArg, ok := tbl.RawGetString("body").(lua.LString)
			if ok {
				reqBody = strings.NewReader(bodyArg.String())
			}

		}

		r, err := http.NewRequest(method, urlStr, reqBody)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LNil)
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 4
		}
		for k, hdrs := range headers {
			for _, v := range hdrs {
				r.Header.Add(k, v)
			}
		}

		// TODO: add cookie(maybe just let lua set it in header themself)

		resp, err := client.Do(r)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LNil)
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 4
		}
		defer resp.Body.Close()
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LNil)
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 4
		}

		respHeaders := L.NewTable()
		for k, vals := range resp.Header {
			tmpHdrs := L.NewTable()
			for _, v := range vals {
				// fmt.Printf("DEBUG: %s %s\n", k, v)
				tmpHdrs.Append(lua.LString(v))
			}
			respHeaders.RawSetString(k, tmpHdrs)
		}

		// respHeaders.ForEach(func(k, v lua.LValue) {
		// 	fmt.Printf("DEBUG: %v %v\n", k.String(), v.String())
		// })

		L.Push(lua.LString(respBody))
		L.Push(lua.LNumber(resp.StatusCode))
		L.Push(respHeaders)
		L.Push(lua.LNil)
		return 4
	}
}

func httpHeadersFromLua(lt *lua.LTable) map[string][]string {
	headerLv := lt.RawGetString("headers")
	if headerLv.Type() == lua.LTNil {
		return map[string][]string{}
	}
	headerTv, ok := headerLv.(*lua.LTable)
	if !ok {
		return map[string][]string{}
	}

	headerMap := make(map[string][]string, 0)

	headerTv.ForEach(func(k, v lua.LValue) {
		if k.Type() != lua.LTString {
			return
		}
		vt, isArr := v.(*lua.LTable)
		if isArr {
			vt.ForEach(func(k2, v2 lua.LValue) {
				if k2.Type() == lua.LTNumber && v2.Type() == lua.LTString {
					headerMap[k.String()] = append(headerMap[k.String()], v2.String())
				}
			})
		} else if v.Type() == lua.LTString {
			headerMap[k.String()] = append(headerMap[k.String()], v.String())
		}
	})

	return headerMap
}
