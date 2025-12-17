import { GetChapterByID } from '$lib/data/api.js';
import { error } from '@sveltejs/kit';

export async function load({ fetch, params,data }) {
    let id = parseInt(params.id)
    if (isNaN(id)) {
        error(400, { message: "Invalid ID" })
    }
    const chap = await GetChapterByID(fetch, id).then(chap => {
        chap.tags.sort((a, b) => a.namespace === b.namespace ? a.label.localeCompare(b.label) : a.namespace.localeCompare(b.namespace))
        return chap
    })

    return {
        chapter: chap
    }
}