export interface Archive{
    id:number
    path:string
    size:number
    modtime:number
    type:string
    partialHash:string
    fileIDs:number[]
}

export interface File{
    id:number
    path:string
    modifiedAt:number
    mime:string
    size:number
}

export interface Page{
    id:number
    name:string
    width:number
    height:number
    isSpred:boolean
    fileID:number
    file:File
}

export interface Tag{
    namespace:string
    label:string
}

export interface Chapter{
    id:number
    name:string
    description:string
    pageIDs:number[]
    coverPageId:number
    tags:Tag[]
}

export interface ChapterSearchResult{
    data:Chapter[]
    total:number
}



export interface Plugin{
    name:string
    description:string
    version:string
    autoRun:boolean
    delay:number
    targetType:string
    config:PluginConfig[]
    param:PluginParam[]
}

export interface PluginConfig{
    name:string
    description:string
    value:string
}

export interface PluginParam extends PluginConfig{

}
export interface ExecPluginRequest{
    name:string
    version:string
    target:number
    param:Record<string,string>
}

