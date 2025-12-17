export function formatBytes(size: number, decimals: number = 2): string {
    if (size === 0) return '0 B';

    const k = 1024; // 1024 bytes = 1KB
    const dm = decimals < 0 ? 0 : decimals; // Ensure decimal places are non-negative
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];

    // Determine the appropriate unit
    const i = Math.floor(Math.log(size) / Math.log(k));

    // Return formatted result with specified decimals
    return parseFloat((size / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}


export function formatTimestamp(milli: number): string {
    const date = new Date(milli)
    return date.toLocaleString()
}

export function parseQuery(query: string): ChapterFilter {
    const strs: string[] = query.match(/(?:[^\\,]|\\.)+/g)?.map(str => str.replace(/\\,/g, ',')) ?? [query]


    const filter = {}as ChapterFilter

    strs.forEach(s => {
        const tagPart = s.match(/(?:[^\\:]|\\.)+/g)
        if (tagPart && tagPart.length >= 2) {
            const namespace = tagPart[0]
            const label = tagPart[1]
            switch (namespace.charAt(0)) {
                case '-':
                    (filter.notHaveTags??=[]).push({
                        namespace:newStringFilter(namespace.substring(1)),
                        label:newStringFilter(label)
                    })
                    break;
                case '~':
                    (filter.containTags??=[]).push({
                        namespace:newStringFilter(namespace.substring(1)),
                        label:newStringFilter(label)
                    })
                    break;
            
                default:
                    (filter.haveTags??=[]).push({
                        namespace:newStringFilter(namespace),
                        label:newStringFilter(label)
                    })
                    break;
            }

        }else{
            switch (s.charAt(0)) {
                case '-':
                    (filter.notHaveNames??=[]).push(newStringFilter(s.substring(1)))
                    break;
                case '~':
                    (filter.containNames??=[]).push(newStringFilter(s.substring(1)))
                    break;
            
                default:
                    (filter.haveNames??=[]).push(newStringFilter(s))
                    break;
            }
        }
    })

    return filter
}

export interface StringFilter {
    type: 'exact' | 'prefix' | 'suffix' | 'infix'
    value:string
}

export interface TagFilter{
    namespace:StringFilter
    label:StringFilter
}

interface ChapterFilter{
    haveNames?:StringFilter[]
    notHaveNames?:StringFilter[]
    containNames?:StringFilter[]
    haveTags?:TagFilter[]
    notHaveTags?:TagFilter[]
    containTags?:TagFilter[]

}

export interface SearchSort{
    by:string,
    descending:boolean,
    namespace?:string
}

export function newStringFilter(s: string):StringFilter {
    const sf={} as StringFilter
    const prefix=s.startsWith("^")
    const suffix=s.endsWith("$")
    if (prefix&&suffix) {
        sf.type='exact'
        sf.value=s.substring(1,s.length-1)
    }else if(prefix){
        sf.type='prefix'
        sf.value=s.substring(1)
    }else if(suffix){
        sf.type='suffix'
        sf.value=s.substring(0,s.length-1)
    }else{
        sf.type='infix'
        sf.value=s
    }

    return sf
}

