import { GetAllChapters, SearchChapter } from '$lib/data/api.js';
import type { ChapterSearchResult } from '$lib/data/model.js';
import { redirect } from '@sveltejs/kit';
import { goto } from '$app/navigation';
import type { SearchSort } from '$lib/data/utils.js';


export async function load({ fetch, url }) {
    let pg = parseInt(url.searchParams.get("p") ?? "")
    if (isNaN(pg)) {
        pg = 1
    }
    let queryPS = parseInt(url.searchParams.get('ps') ?? '')
    if (isNaN(queryPS)) {
        queryPS = 10
    }
    const sordDesc = url.searchParams.get('sortDesc') == 'true'
    let sort: SearchSort | null = null
    const sortQ=url.searchParams.get('sort')??''
    switch (sortQ) {
        case '':
            break;
        case 'id':
            sort = {
                by: 'id',
                descending: sordDesc
            }
            break;
        case 'name':
            sort = {
                by: 'name',
                descending: sordDesc
            }
            break;
        default:
            sort = {
                by: 'tagnamespace',
                descending: sordDesc,
                namespace:sortQ
            }
            break;
    }


    const res: ChapterSearchResult = await SearchChapter(fetch,
        pg,
        queryPS, url.searchParams.get("q") ?? '',
        sort?[sort]:undefined
    );
    res.data.forEach(chap => {
        chap.tags.sort((a, b) => a.namespace === b.namespace ? a.label.localeCompare(b.label) : a.namespace.localeCompare(b.namespace))
    })
    return {
        chapters: res.data,
        totalChapter: res.total,
        perPage: queryPS
    }
}