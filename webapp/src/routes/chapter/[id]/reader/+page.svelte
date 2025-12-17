<script lang="ts">
	import { preloadCode, preloadData, replaceState } from '$app/navigation';
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Pagination from '$lib/components/ui/pagination/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { GetChapterByID, PageURL } from '$lib/data/api';
	import { onMount } from 'svelte';

	let { data } = $props();

	const pageURLs = data.chapter.pageIDs.map((p) => PageURL(p));
	let currentPage = $state(1);
	onMount(() => {
		const handle = (e: KeyboardEvent) => {
			switch (e.key) {
				case 'A':
				case 'a':
				case 'ArrowRight':
				case ' ':
					currentPage++;
					break;
				case 'D':
				case 'd':
				case 'ArrowLeft':
					currentPage--;
					break;

				default:
					break;
			}
		};
		document.addEventListener('keydown', handle);
		return () => {
			document.removeEventListener('keydown', handle);
		};
	});

	$effect(() => {
		let hash = page.url.hash.substring(1);
		let pg = parseInt(hash);
		if (!isNaN(pg) && pg > 1 && pg <= data.chapter.pageIDs.length) {
			currentPage = pg;
		} else if (pg > data.chapter.pageIDs.length) {
			currentPage = data.chapter.pageIDs.length;
		}
	});
	$effect(() => {
		if (page.url.hash != `#${currentPage}`) {
			replaceState(`#${currentPage}`, page.state);
		}
	});

	let currentPageURL = $derived(pageURLs[currentPage - 1]);
	const imageRefMap: Map<string, boolean> = new Map();
	$effect(() => {
		const urls = pageURLs.slice(
			Math.max(0, currentPage - 1 - 5),
			Math.min(pageURLs.length - 1, currentPage + 5)
		);
		urls.forEach((u) => {
			if (u == currentPageURL) {
				return;
			}
			let ok = imageRefMap.get(u);
			if (!ok) {
				const el = new Image();
				el.src = u;
				imageRefMap.set(u, true);
			}
		});
	});

	let readerContainer: HTMLDivElement | null = $state(null);
	let fitToHeight = $state(false);

	let isInFullscreen = $state(false);

	function enableFullscreen() {
		readerContainer?.requestFullscreen().then(() => {
			isInFullscreen = true;
		});
	}
</script>

<div
	class="relative flex items-center justify-center overflow-auto {isInFullscreen
		? ''
		: 'px-2 py-4'}"
	bind:this={readerContainer}
>
	<img
		class="{fitToHeight ? 'max-h-screen' : ''} max-w-full"
		alt="page{currentPage}"
		src={currentPageURL}
	/>

	{#if currentPage > 1}
		<a
			onclick={(e) => {
				e.preventDefault();
				currentPage--;
			}}
			class="absolute top-0 bottom-0 left-0 z-1 h-full w-1/2"
			tabindex={-1}
			href={`/chapter/${data.chapter.id}/reader#${currentPage - 1}`}
			aria-label="previous page"
		></a>
	{/if}
	{#if currentPage < data.chapter.pageIDs.length}
		<a
			onclick={(e) => {
				e.preventDefault();
				currentPage++;
			}}
			class="absolute top-0 right-0 bottom-0 z-1 h-full w-1/2"
			role="button"
			tabindex={-1}
			href={`/chapter/${data.chapter.id}/reader#${currentPage + 1}`}
			aria-label="next page"
		></a>
	{/if}
</div>
<Pagination.Root bind:page={currentPage} count={pageURLs.length} perPage={1}>
	{#snippet children({ pages })}
		<Pagination.Content>
			<Pagination.Item>
				<Pagination.PrevButton />
			</Pagination.Item>

			<Select.Root
				type="single"
				onValueChange={(v) => {
					const pg = parseInt(v);
					if (pg >= 1 && pg <= data.chapter.pageIDs.length) {
						currentPage = pg;
					}
				}}
			>
				<Select.Trigger class=""
					>Page {currentPage} of {data.chapter.pageIDs.length}
				</Select.Trigger>
				<Select.Content>
					{#each pageURLs as url, idx}
						<Select.Item value={(idx + 1).toString()}>Page {(idx + 1).toString()}</Select.Item>
					{/each}
				</Select.Content>
			</Select.Root>
			<Pagination.Item>
				<Pagination.NextButton />
			</Pagination.Item>
		</Pagination.Content>
	{/snippet}
</Pagination.Root>

<div class="mx-auto my-2 w-fit">
	<button onclick={enableFullscreen}>Fullscreen</button>|
	<a href="/chapter/{data.chapter.id}">Go Back to Chapter</a>
</div>
