import { GetAllPlugins } from '$lib/data/api.js'

export async function load({fetch}) {
    const plugs = await GetAllPlugins(fetch)
    return {
        plugins:plugs
    }
}