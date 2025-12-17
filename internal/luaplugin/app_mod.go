package luaplugin

import (
	"bytes"

	"github.com/Ollinar/scuff/internal/app"
	"github.com/Ollinar/scuff/internal/model"

	lua "github.com/yuin/gopher-lua"
)

func WithAppModlue(app *app.App) LuaModule {
	return func(L *lua.LState) {
		L.PreloadModule("app", appModuleLoader(app))
	}
}

func appModuleLoader(app *app.App) lua.LGFunction {
	return func(L *lua.LState) int {
		appModFunc := map[string]lua.LGFunction{
			"getChaptersByIds":      getChaptersByIds(app),
			"addChapterFromArchive": addChapterFromArchive(app),
			"updateChapter":         updateChapter(app),
			"getPagesByIds":         getPagesByIds(app),
			"getFilesByIds":         getFilesByIds(app),
			"getArchivesByIds":      getArchivesByIds(app),
			"readFile":              readFile(app),
		}
		mod := L.SetFuncs(L.NewTable(), appModFunc)

		L.Push(mod)
		return 1
	}
}

func getPagesByIds(app *app.App) lua.LGFunction {
	return func(L *lua.LState) int {
		ids := make([]model.ID, 0, L.GetTop()-1)

		for i := 1; i <= L.GetTop(); i++ {
			id := L.CheckNumber(i)
			ids = append(ids, model.ID(id))
		}
		pgs, err := app.Page().GetByIDs(L.Context(), ids)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		tb := L.NewTable()
		for _, pg := range pgs {
			ud := L.NewUserData()
			ud.Value = &pg
			L.SetMetatable(ud, L.GetTypeMetatable(PageTypeKey))
			tb.Append(ud)
		}
		L.Push(tb)
		L.Push(lua.LNil)
		return 2
	}
}

func getArchivesByIds(app *app.App) lua.LGFunction {
	return func(L *lua.LState) int {
		ids := make([]model.ID, 0, L.GetTop()-1)

		for i := 1; i <= L.GetTop(); i++ {
			id := L.CheckNumber(i)
			ids = append(ids, model.ID(id))
		}
		arcs, err := app.Archive().GetByIDs(L.Context(), ids)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		tb := L.NewTable()
		for _, arc := range arcs {
			ud := L.NewUserData()
			ud.Value = &arc
			L.SetMetatable(ud, L.GetTypeMetatable(ArchiveTypeKey))
			tb.Append(ud)
		}
		L.Push(tb)
		L.Push(lua.LNil)
		return 2
	}
}

func getFilesByIds(app *app.App) lua.LGFunction {
	return func(L *lua.LState) int {
		ids := make([]model.ID, 0, L.GetTop()-1)

		for i := 1; i <= L.GetTop(); i++ {
			id := L.CheckNumber(i)
			ids = append(ids, model.ID(id))
		}
		fls, err := app.Archive().GetFilesByIDs(L.Context(), ids)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		tb := L.NewTable()
		for _, fl := range fls {
			ud := L.NewUserData()
			ud.Value = &fl
			L.SetMetatable(ud, L.GetTypeMetatable(FileTypeKey))
			tb.Append(ud)
		}
		L.Push(tb)
		L.Push(lua.LNil)
		return 2
	}
}

func getChaptersByIds(app *app.App) lua.LGFunction {
	return func(L *lua.LState) int {
		ids := make([]model.ID, 0, L.GetTop()-1)

		for i := 1; i <= L.GetTop(); i++ {
			id := L.CheckNumber(i)
			ids = append(ids, model.ID(id))
		}
		chaps, err := app.Chapter().GetByIDs(L.Context(), ids)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		tb := L.NewTable()
		for _, chp := range chaps {
			ud := L.NewUserData()
			ud.Value = &chp
			L.SetMetatable(ud, L.GetTypeMetatable(ChapterTypeKey))
			tb.Append(ud)
		}
		L.Push(tb)
		L.Push(lua.LNil)
		return 2
	}
}

func addChapterFromArchive(app *app.App) lua.LGFunction {
	return func(L *lua.LState) int {
		arcID := L.CheckNumber(1)
		chp, err := app.Chapter().AddChapterFromArchive(L.Context(), model.ID(arcID))
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		ud := L.NewUserData()
		ud.Value = &chp
		L.SetMetatable(ud, L.GetTypeMetatable(ChapterTypeKey))
		L.Push(ud)
		L.Push(lua.LNil)
		return 2
	}
}

func updateChapter(app *app.App) lua.LGFunction {
	return func(L *lua.LState) int {
		chapID := L.CheckNumber(1)
		chap := checkChapterModel(L, 2)

		newCh, err := app.Chapter().Update(L.Context(), model.ID(chapID), *chap)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		ud := L.NewUserData()
		ud.Value = &newCh
		L.SetMetatable(ud, L.GetTypeMetatable(ChapterTypeKey))
		L.Push(ud)
		L.Push(lua.LNil)
		return 2
	}
}

func readFile(app *app.App) lua.LGFunction {
	return func(L *lua.LState) int {
		fl := checkFileModel(L, 1)
		bBuf := make([]byte, 0, fl.Size)
		buf := bytes.NewBuffer(bBuf)
		err := app.File().ReadFile(L.Context(), buf, *fl)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LString(buf.String()))
		L.Push(lua.LNil)
		return 2
	}
}
