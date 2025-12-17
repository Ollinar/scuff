import { GetAllPlugins } from '$lib/data/api.js'

export async function load({fetch}) {
    let plugs = await GetAllPlugins(fetch)
    plugs=plugs.filter(p=>p.targetType=="chapter")
    return {
        plugins:plugs
    }
}