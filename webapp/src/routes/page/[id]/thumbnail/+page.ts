import { error, redirect } from "@sveltejs/kit"
import type { PageLoad } from "./$types"
import { env } from "$env/dynamic/public"
import { goto } from "$app/navigation"

export const load = async ({ params, url,fetch }) => {

    let id = parseInt(params.id)
    if (isNaN(id)) {
        error(400, { message: "Invalid ID" })
    }
    const apiUrl = new URL(`${env.PUBLIC_API_ENDPOINT}/api/v1/page/${id}/thumbnail/`)
    url.searchParams.forEach((k, v) => apiUrl.searchParams.set(k, v))

    let res=await fetch(apiUrl)

    let blob = await res.blob()
    let blobUrl = URL.createObjectURL(blob)
    goto(apiUrl)
    // redirect(308,blobUrl)
}