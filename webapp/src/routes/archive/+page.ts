import { GetAllArchives } from "$lib/data/api";
import type { Archive } from "$lib/data/model";

export async function load({ fetch }) {
        let data =GetAllArchives(fetch)
        data.catch(()=>{})
        return {
        archives:data 
    }

}