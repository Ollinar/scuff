import { GetArchiveByID, GetArchivePages } from "$lib/data/api"
import { error } from "@sveltejs/kit"


export async function load({ params, fetch }) {
    let id = parseInt(params.id)
    if (isNaN(id)) {
        error(400, { message: "Invalid ID" })
    }


    const archive = GetArchiveByID(fetch, id)
    const pages = GetArchivePages(fetch, id)
    return {
        archive: await archive,
        pages: pages
    }
}