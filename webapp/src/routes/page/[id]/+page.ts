import { error, redirect } from "@sveltejs/kit"
import type { PageLoad } from "./$types"
import { env } from "$env/dynamic/public"

export const load: PageLoad = async ({ params, url }) => {

    let id = parseInt(params.id)
    if (isNaN(id)) {
        error(400, { message: "Invalid ID" })
    }
    const apiUrl = new URL(`${env.PUBLIC_API_ENDPOINT}/api/v1/page/${id}/`)
    url.searchParams.forEach((k, v) => apiUrl.searchParams.set(k, v))

    redirect(308, apiUrl)

}