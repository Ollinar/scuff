import { PUBLIC_API_URL } from "$env/static/public";
import { error, json } from "@sveltejs/kit";
import type { Archive, Page, Chapter, ChapterSearchResult, Plugin, PluginConfig, ExecPluginRequest } from "./model";
import { newStringFilter, parseQuery, type SearchSort } from "./utils";


export async function GetAllArchives(fetch: typeof globalThis.fetch): Promise<Archive[]> {
    try {
        const res = await fetch(`${PUBLIC_API_URL}/api/v1/archive/`)
        if (!res.ok) {
            error(res.status, res.statusText)
        }
        const data: Archive[] = await res.json();
        return data

    } catch (err) {
        error(500, err instanceof Error ? err.message : "")
    }
}

export async function GetArchiveByID(fetch: typeof globalThis.fetch, id: number): Promise<Archive> {
    try {
        const res = await fetch(`${PUBLIC_API_URL}/api/v1/archive/${id}`)
        if (!res.ok) {
            error(res.status, res.statusText)
        }
        const data: Archive = await res.json();
        return data
    } catch (err) {
        error(500, err instanceof Error ? err.message : "")
    }
}

export async function ScanContentDir(fetch: typeof globalThis.fetch) {
    try {
        const res = await fetch(`${PUBLIC_API_URL}/api/v1/archive/scan`, {
            method: "POST",
        })
        if (!res.ok) {
            error(res.status, res.statusText)
        }
        return
    } catch (err) {
        error(500, err instanceof Error ? err.message : "")
    }

}

export async function GenratePagesFromArchive(id: number): Promise<Page[]> {

    try {
        let res = await fetch(PUBLIC_API_URL + `/api/v1/archive/${id}/page/generate`, {
            method: "POST"
        })
        if (!res.ok) {
            error(res.status, res.statusText)
        }
        const data: Page[] = await res.json()
        return data
    } catch (err) {
        error(500, err instanceof Error ? err.message : "")
    }
}

export async function GetArchivePages(fetch: typeof globalThis.fetch, id: number): Promise<Page[]> {
    try {
        let res = await fetch(`${PUBLIC_API_URL}/api/v1/archive/${id}/page`)
        if (!res.ok) {
            throw new Error(res.statusText);
        }
        let data = await res.json()
        return data
    } catch (error) {
        if (error instanceof Error) {
            return new Promise<Page[]>((resolve, reject) => {
                reject(error.message)
            })
        }
        return new Promise<Page[]>((resolve, reject) => {
            reject(error)
        })
    }
}

export async function GenrateChapterFromArchive(id: number): Promise<Chapter> {
    try {
        let res = await fetch(PUBLIC_API_URL + `/api/v1/archive/${id}/chapter/generate`, {
            method: "POST"
        })
        if (!res.ok) {
            error(res.status, res.statusText)
        }
        let data: Chapter = await res.json()
        return data
    } catch (err) {
        error(500, err instanceof Error ? err.message : "")
    }
}


export function PageTumbnailURL(id: number): string {
    return `${PUBLIC_API_URL}/api/v1/page/${id}/thumbnail/?width=500&format=webp`
}

export function PageURL(id: number): string {
    return `${PUBLIC_API_URL}/api/v1/page/${id}/?format=webp`
}


export async function GetAllChapters(fetch: typeof globalThis.fetch): Promise<Chapter[]> {
    try {
        const res = await fetch(`${PUBLIC_API_URL}/api/v1/chapter/`);
        if (!res.ok) {
            error(res.status, res.statusText);
        }
        const data: Chapter[] = await res.json();
        return data
    } catch (err) {
        error(500, err instanceof Error ? err.message : "");
    }
}

export async function GetChapterByID(fetch: typeof globalThis.fetch, id: number): Promise<Chapter> {

    try {
        const res = await fetch(`${PUBLIC_API_URL}/api/v1/chapter/${id}/`);
        if (!res.ok) {
            error(res.status, res.statusText);
        }
        const data: Chapter = await res.json();
        return data
    } catch (err) {
        error(500, err instanceof Error ? err.message : "");
    }
}
export async function SearchChapter(fetch: typeof globalThis.fetch, page: number, pageSize: number, query: string,sort?:SearchSort[]): Promise<ChapterSearchResult> {

    const filter = parseQuery(query)
    try {
        let req: Record<string, any> = { ...filter }
        req.page = page
        req.pageSize = pageSize
        req.sorting=[
            {
                by:"id",
            }
        ]

        if (sort) {
            req.sorting=[]
            sort.forEach(s=>{
                req.sorting.push({
                    by:s.by,
                    descending:s.descending,
                    namespace:s.namespace?newStringFilter(s.namespace):null
                })
            })
        }

        const res = await fetch(`${PUBLIC_API_URL}/api/v1/chapter/search`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(req)
        });
        if (!res.ok) {
            error(res.status, res.statusText);
        }
        const data: ChapterSearchResult = await res.json();
        return data
    } catch (err) {
        error(500, err instanceof Error ? err.message : "");
    }
}

export async function PatchChapter(fetch: typeof globalThis.fetch, id: number, chap: Partial<Chapter>): Promise<Chapter> {

    try {
        const res = await fetch(`${PUBLIC_API_URL}/api/v1/chapter/${id}/`, {
            method: 'PATCH',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(chap)
        });
        if (!res.ok) {
            error(res.status, res.statusText)
        }
        const data: Chapter = await res.json()
        return data

    } catch (err) {
        error(500, err instanceof Error ? err.message : "")
    }
}
export async function DeleteChapter(fetch: typeof globalThis.fetch, id: number): Promise<void> {

    try {
        const res = await fetch(`${PUBLIC_API_URL}/api/v1/chapter/${id}/`, {
            method: 'DELETE',
        });
        if (!res.ok) {
            error(res.status, res.statusText)
        }
        return

    } catch (err) {
        error(500, err instanceof Error ? err.message : "")
    }
}

export const PostPluginEndpoint = `${PUBLIC_API_URL}/api/v1/plugin/`

export async function PostPlugin(fetch: typeof globalThis.fetch, fd: FormData): Promise<void> {
    try {
        const res = await fetch(`${PUBLIC_API_URL}/api/v1/plugin/upload`, {
            method: 'POST',
            body: fd

        });
        if (!res.ok) {
            error(res.status, res.statusText)
        }

        return
    } catch (err) {
        error(500, err instanceof Error ? err.message : '')
    }

}

export async function GetAllPlugins(fetch: typeof globalThis.fetch): Promise<Plugin[]> {
    try {
        const res = await fetch(`${PUBLIC_API_URL}/api/v1/plugin/`)
        if (!res.ok) {
            error(res.status, res.statusText)
        }
        const data: Plugin[] = await res.json()
        return data
    } catch (err) {
        error(500, err instanceof Error ? err.message : '')
    }
}

export async function UpdatePluginConfig(fetch: typeof globalThis.fetch, plugin: Plugin): Promise<void> {
    const configs: Record<string, any> = {}

    plugin.config.forEach(c => {
        configs[c.name] = c.value
    })
    const pyload = {
        name: plugin.name,
        version: plugin.version,
        autoRun: plugin.autoRun,
        delay: plugin.delay,
        config: configs
    }
    try {
        const res = await fetch(`${PUBLIC_API_URL}/api/v1/plugin/`, {
            method: 'PATCH',
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(pyload)

        });
        if (!res.ok) {
            error(res.status, res.statusText)
        }

        return
    } catch (err) {
        error(500, err instanceof Error ? err.message : '')
    }
}

export async function ExecutePlugin(fetch: typeof globalThis.fetch,target:number, plugin: Plugin): Promise<void> {
    const params: Record<string, string> = {}

    plugin.param.forEach(c => {
        params[c.name] = c.value.toString()
    })
    const pyload:ExecPluginRequest = {
        name: plugin.name,
        version: plugin.version,
        target:target,
        param:params
    }
    try {
        const res = await fetch(`${PUBLIC_API_URL}/api/v1/plugin/`, {
            method: 'POST',
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(pyload)

        });
        if (!res.ok) {
            error(res.status, res.statusText)
        }

        return
    } catch (err) {
        error(500, err instanceof Error ? err.message : '')
    }
}