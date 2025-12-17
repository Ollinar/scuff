package luaplugin

import (
	"github.com/Ollinar/scuff/internal/model"

	lua "github.com/yuin/gopher-lua"
)

const (
	FileTypeKey    string = "appmodule.file"
	ArchiveTypeKey string = "appmodule.archive"
	TagTypeKey     string = "appmodule.tag"
	PageTypeKey    string = "appmodule.page"
	ChapterTypeKey string = "appmodule.chapter"
)

func loadTypes(L *lua.LState) *lua.LTable {
	typesTbl := L.NewTable()

	fileModelType := L.NewTypeMetatable(FileTypeKey)
	L.SetField(fileModelType, "__index", L.SetFuncs(L.NewTable(), fileTypeMethods))
	L.SetField(typesTbl, "File", fileModelType)

	archiveModelType := L.NewTypeMetatable(ArchiveTypeKey)
	L.SetField(archiveModelType, "__index", L.SetFuncs(L.NewTable(), archiveTypeMethods))
	L.SetField(typesTbl, "Archive", archiveModelType)

	tagModelType := L.NewTypeMetatable(TagTypeKey)
	L.SetField(tagModelType, "__index", L.SetFuncs(L.NewTable(), tagTypeMethods))
	L.SetField(tagModelType, "new", L.NewFunction(newTagModel))
	L.SetField(typesTbl, "Tag", tagModelType)

	pageModelType := L.NewTypeMetatable(PageTypeKey)
	L.SetField(pageModelType, "__index", L.SetFuncs(L.NewTable(), pageTypeMethods))
	L.SetField(pageModelType, "new", L.NewFunction(newPageModel))
	L.SetField(typesTbl, "Page", pageModelType)

	chapterType := L.NewTypeMetatable(ChapterTypeKey)
	L.SetField(chapterType, "__index", L.SetFuncs(L.NewTable(), chapterTypeMethods))
	L.SetField(chapterType, "new", L.NewFunction(newChapterModel))
	L.SetField(typesTbl, "Chapter", chapterType)
	return typesTbl
}

var chapterTypeMethods = map[string]lua.LGFunction{
	"name": func(L *lua.LState) int {
		chap := checkChapterModel(L, 1)
		if L.GetTop() == 2 {
			chap.Name = L.CheckString(2)
			return 0
		}
		L.Push(lua.LString(chap.Name))
		return 1
	},
	"description": func(L *lua.LState) int {
		chap := checkChapterModel(L, 1)
		if L.GetTop() == 2 {
			chap.Descripion = L.CheckString(2)
			return 0
		}
		L.Push(lua.LString(chap.Descripion))
		return 1
	},
	"id": func(L *lua.LState) int {
		chap := checkChapterModel(L, 1)
		if L.GetTop() == 2 {
			chap.ID = model.ID(L.CheckNumber(2))
			return 0
		}
		L.Push(lua.LNumber(chap.ID))
		return 1
	},
	"coverPageId": func(L *lua.LState) int {
		chap := checkChapterModel(L, 1)
		if L.GetTop() == 2 {
			chap.CoverPageID = model.ID(L.CheckNumber(2))
			return 0
		}
		L.Push(lua.LNumber(chap.CoverPageID))
		return 1
	},
	"tags": func(L *lua.LState) int {
		chap := checkChapterModel(L, 1)
		if L.GetTop() == 2 {
			args := L.CheckTable(2)
			ok := true
			tags := make([]model.Tag, 0, args.Len())
			args.ForEach(func(k, v lua.LValue) {
				if !ok {
					return
				}
				ud, isUD := v.(*lua.LUserData)
				if k.Type() != lua.LTNumber || !isUD {
					ok = false
					return
				}
				tg, isTg := ud.Value.(*model.Tag)
				if !isTg {
					ok = false
					return
				}
				tags = append(tags, *tg)
			})
			if !ok {
				L.ArgError(2, "expected array of tags")
				return 0
			}
			chap.Tags = tags
			return 0
		}
		lt := L.NewTable()
		for _, v := range chap.Tags {
			ud := L.NewUserData()
			ud.Value = &v
			L.SetMetatable(ud, L.GetTypeMetatable(TagTypeKey))
			lt.Append(ud)
		}
		L.Push(lt)
		return 1
	},
	"pageIds": func(L *lua.LState) int {
		chap := checkChapterModel(L, 1)
		if L.GetTop() == 2 {
			args := L.CheckTable(2)
			ok := true
			ids := make([]model.ID, 0, args.Len())
			args.ForEach(func(k, v lua.LValue) {
				if !ok {
					return
				}
				id := lua.LVAsNumber(v)
				if k.Type() != lua.LTNumber || (v.Type() != lua.LTNumber && id == 0) {
					ok = false
					return
				}
				ids = append(ids, model.ID(id))
			})
			if !ok {
				L.ArgError(2, "expected array of numbers")
				return 0
			}
			chap.PageIDs = ids
			return 0
		}
		lt := L.NewTable()
		for _, v := range chap.PageIDs {
			lt.Append(lua.LNumber(v))
		}
		L.Push(lt)
		return 1
	},
}

var pageTypeMethods = map[string]lua.LGFunction{
	"id": func(L *lua.LState) int {
		pg := checkPageModel(L, 1)
		if L.GetTop() == 2 {
			pg.ID = model.ID(L.CheckNumber(2))
			return 0
		}
		L.Push(lua.LNumber(pg.ID))
		return 1
	},
	"fileId": func(L *lua.LState) int {
		pg := checkPageModel(L, 1)
		if L.GetTop() == 2 {
			pg.FileID = model.ID(L.CheckNumber(2))
			return 0
		}
		L.Push(lua.LNumber(pg.FileID))
		return 1
	},
	"isSpread": func(L *lua.LState) int {
		pg := checkPageModel(L, 1)
		if L.GetTop() == 2 {
			pg.IsSpread = L.CheckBool(2)
			return 0
		}
		L.Push(lua.LBool(pg.IsSpread))
		return 1
	},
	"name": func(L *lua.LState) int {
		pg := checkPageModel(L, 1)
		if L.GetTop() == 2 {
			pg.Name = L.CheckString(2)
			return 0
		}
		L.Push(lua.LString(pg.Name))
		return 1
	},
	"width": func(L *lua.LState) int {
		pg := checkPageModel(L, 1)
		if L.GetTop() == 2 {
			pg.Width = L.CheckInt(2)
			return 0
		}
		L.Push(lua.LNumber(pg.Width))
		return 1
	},
	"height": func(L *lua.LState) int {
		pg := checkPageModel(L, 1)
		if L.GetTop() == 2 {
			pg.Height = L.CheckInt(2)
			return 0
		}
		L.Push(lua.LNumber(pg.Height))
		return 1
	},
	"file": func(L *lua.LState) int {
		pg := checkPageModel(L, 1)
		if L.GetTop() == 2 {
			fl := checkFileModel(L, 2)
			pg.File = *fl
			return 0
		}
		ud := L.NewUserData()
		ud.Value = &pg.File
		L.SetMetatable(ud, L.GetTypeMetatable(FileTypeKey))
		L.Push(ud)
		return 1
	},
}

var tagTypeMethods = map[string]lua.LGFunction{
	"namespace": func(L *lua.LState) int {
		tg := checkTagModel(L, 1)
		if L.GetTop() == 2 {
			tg.Namespace = L.CheckString(2)
			return 0
		}
		L.Push(lua.LString(tg.Namespace))
		return 1
	},
	"label": func(L *lua.LState) int {
		tg := checkTagModel(L, 1)
		if L.GetTop() == 2 {
			tg.Label = L.CheckString(2)
			return 0
		}
		L.Push(lua.LString(tg.Label))
		return 1
	},
}

var archiveTypeMethods = map[string]lua.LGFunction{
	"id": func(L *lua.LState) int {
		arc := checkArchiveModel(L, 1)
		if L.GetTop() == 2 {
			arc.ID = model.ID(L.CheckNumber(2))
			return 0
		}
		L.Push(lua.LNumber(arc.ID))
		return 1
	},
	"size": func(L *lua.LState) int {
		arc := checkArchiveModel(L, 1)
		if L.GetTop() == 2 {
			arc.Size = L.CheckInt64(2)
			return 0
		}
		L.Push(lua.LNumber(arc.Size))
		return 1
	},
	"modtime": func(L *lua.LState) int {
		arc := checkArchiveModel(L, 1)
		if L.GetTop() == 2 {
			arc.ModTime = L.CheckInt64(2)
			return 0
		}
		L.Push(lua.LNumber(arc.ModTime))
		return 1
	},
	"path": func(L *lua.LState) int {
		arc := checkArchiveModel(L, 1)
		if L.GetTop() == 2 {
			arc.Path = L.CheckString(2)
			return 0
		}
		L.Push(lua.LString(arc.Path))
		return 1
	},
	"type": func(L *lua.LState) int {
		arc := checkArchiveModel(L, 1)
		if L.GetTop() == 2 {
			arc.Type = L.CheckString(2)
			return 0
		}
		L.Push(lua.LString(arc.Type))
		return 1
	},
	"partialHash": func(L *lua.LState) int {
		arc := checkArchiveModel(L, 1)
		if L.GetTop() == 2 {
			arc.PartialHash = L.CheckString(2)
			return 0
		}
		L.Push(lua.LString(arc.PartialHash))
		return 1
	},
	"fileIds": func(L *lua.LState) int {
		arc := checkArchiveModel(L, 1)
		if L.GetTop() == 2 {
			args := L.CheckTable(2)
			ok := true
			ids := make([]model.ID, 0, args.Len())
			args.ForEach(func(k, v lua.LValue) {
				if !ok {
					return
				}
				id := lua.LVAsNumber(v)
				if k.Type() != lua.LTNumber || (v.Type() != lua.LTNumber && id == 0) {
					ok = false
					return
				}
				ids = append(ids, model.ID(id))
			})
			if !ok {
				L.ArgError(2, "expected array of numbers")
				return 0
			}
			arc.FileIds = ids
			return 0
		}
		lt := L.NewTable()
		for _, v := range arc.FileIds {
			lt.Append(lua.LNumber(v))
		}
		L.Push(lt)
		return 1
	},
}

var fileTypeMethods = map[string]lua.LGFunction{
	"id": func(L *lua.LState) int {
		fl := checkFileModel(L, 1)
		if L.GetTop() == 2 {
			fl.ID = model.ID(L.CheckNumber(2))
			return 0
		}
		L.Push(lua.LNumber(fl.ID))
		return 1
	},
	"archiveId": func(L *lua.LState) int {
		fl := checkFileModel(L, 1)
		if L.GetTop() == 2 {
			fl.ArchiveID = model.ID(L.CheckNumber(2))
			return 0
		}
		L.Push(lua.LNumber(fl.ArchiveID))
		return 1
	},
	"modtime": func(L *lua.LState) int {
		fl := checkFileModel(L, 1)
		if L.GetTop() == 2 {
			fl.ModTime = L.CheckInt64(2)
			return 0
		}
		L.Push(lua.LNumber(fl.ModTime))
		return 1
	},
	"size": func(L *lua.LState) int {
		fl := checkFileModel(L, 1)
		if L.GetTop() == 2 {
			fl.Size = L.CheckInt64(2)
			return 0
		}
		L.Push(lua.LNumber(fl.Size))
		return 1
	},
	"archivePath": func(L *lua.LState) int {
		fl := checkFileModel(L, 1)
		if L.GetTop() == 2 {
			fl.ArchivePath = L.CheckString(2)
			return 0
		}
		L.Push(lua.LString(fl.ArchivePath))
		return 1
	},
	"path": func(L *lua.LState) int {
		fl := checkFileModel(L, 1)
		if L.GetTop() == 2 {
			fl.Path = L.CheckString(2)
			return 0
		}
		L.Push(lua.LString(fl.Path))
		return 1
	},
	"mime": func(L *lua.LState) int {
		fl := checkFileModel(L, 1)
		if L.GetTop() == 2 {
			fl.Mime = L.CheckString(2)
			return 0
		}
		L.Push(lua.LString(fl.Mime))
		return 1
	},
}

func checkChapterModel(L *lua.LState, n int) *model.Chapter {
	ud := L.CheckUserData(n)
	if ch, ok := ud.Value.(*model.Chapter); ok {
		return ch
	}
	L.ArgError(n, "expected chapter")
	return nil
}

func checkPageModel(L *lua.LState, n int) *model.Page {
	ud := L.CheckUserData(n)
	if pg, ok := ud.Value.(*model.Page); ok {
		return pg
	}
	L.ArgError(n, "expected page")
	return nil
}

func checkTagModel(L *lua.LState, n int) *model.Tag {
	ud := L.CheckUserData(n)
	if tg, ok := ud.Value.(*model.Tag); ok {
		return tg
	}
	L.ArgError(n, "expected tag")
	return nil
}

func checkArchiveModel(L *lua.LState, n int) *model.Archive {
	ud := L.CheckUserData(n)
	if arc, ok := ud.Value.(*model.Archive); ok {
		return arc
	}
	L.ArgError(n, "expected archive")
	return nil
}

func checkFileModel(L *lua.LState, n int) *model.File {
	ud := L.CheckUserData(n)
	if fl, ok := ud.Value.(*model.File); ok {
		return fl
	}
	L.ArgError(n, "expected file")
	return nil
}

func newTagModel(L *lua.LState) int {
	tg := &model.Tag{Namespace: L.CheckString(1), Label: L.CheckString(2)}
	ud := L.NewUserData()
	ud.Value = tg
	L.SetMetatable(ud, L.GetTypeMetatable(TagTypeKey))
	L.Push(ud)
	return 1
}

func newPageModel(L *lua.LState) int {
	ud := L.NewUserData()
	ud.Value = &model.Page{}

	L.SetMetatable(ud, L.GetTypeMetatable(PageTypeKey))
	L.Push(ud)
	return 1
}

func newChapterModel(L *lua.LState) int {
	ud := L.NewUserData()
	ud.Value = &model.Chapter{}
	L.SetMetatable(ud, L.GetTypeMetatable(ChapterTypeKey))
	L.Push(ud)

	return 1
}
