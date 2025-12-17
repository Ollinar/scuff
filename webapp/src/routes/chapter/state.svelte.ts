import { browser } from "$app/environment";

export function initialChapterPageSize() {
    const stored = browser ? localStorage.getItem("chapter-per-page") : null;
    const ps = parseInt(stored ?? '');
    let s = $state({
        pageSize: isNaN(ps) ? 10 : ps
    })
    $effect.root(() => {
        $effect(() => {
            if (browser) {
                localStorage.setItem("chapter-per-page", s.pageSize.toString())
            }
        })
    })

    return s
}

// export const chapterPageSize = initialChapterPageSize()




