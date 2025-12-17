<script lang="ts">
	import { page } from '$app/state';
	import { PageThumbnail } from '$lib/components/ui/page-thumbnail/index.js';
	import { toast } from 'svelte-sonner';
	import type { Archive, Page, Chapter } from '$lib/data/model.js';
	import { PageTumbnailURL } from '$lib/data/api.js';
	import { formatBytes } from '$lib/data/utils.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import EllipsisIcon from '@lucide/svelte/icons/ellipsis';
	import { Button } from '$lib/components/ui/button/index.js';

	let { data } = $props();

	let pages: (Page & { selected: boolean })[] = $state([]);

	let pageMode: 'select' | 'navigate' | 'display' = $state('display');
	let selectedCount = $derived(pages.filter((v) => v.selected).length);

	$effect(() => {
		toast.promise(data.pages, {
			loading: 'Fetching Pages...',

			success: (data) => {
				pages = data.map((p) => {
					return { ...p, selected: false };
				});
				return data.length + ' page fetched';
			},
			error: (err) => 'Failed to get pages ' + err,
			position: 'top-left',
			closeButton: true,
			richColors: true
		});
	});
</script>

<div class="m-4 flex flex-wrap border-2 p-2">
	<div class="m-4 grow">
		<h1 class="">
			{data.archive.path}
		</h1>
		<h6>
			{formatBytes(data.archive.size)}
		</h6>
		<h6>
			{`${pages.length} Pages`}
		</h6>
	</div>
	<div class=" flex flex-col md:items-end">
		<div class="">
			<DropdownMenu.Root>
				<DropdownMenu.Trigger>
					{#snippet child({ props })}
						<Button {...props} variant="ghost" size="icon" class="relative size-8 p-0">
							<span class="sr-only">Open menu</span>
							<EllipsisIcon />
						</Button>
					{/snippet}
				</DropdownMenu.Trigger>
				{#if pageMode=='select'}
				<DropdownMenu.Content>
					<DropdownMenu.Item onclick={() => {

					}} disabled={selectedCount==0}>Add To Chapter</DropdownMenu.Item>
				</DropdownMenu.Content>
					
				{:else}
				<DropdownMenu.Content>
					<DropdownMenu.Item onclick={() => (pageMode = 'select')}>Select Pages</DropdownMenu.Item>
				</DropdownMenu.Content>
				{/if}
			</DropdownMenu.Root>
		</div>
		{#if pageMode === 'select'}
			<div>
				<p>{`${selectedCount} of ${pages.length} selected`}</p>
			</div>
			<div class="felx right-3 bottom-3">
				<Button
					onclick={() => {
						pages.forEach((p) => (!p.selected ? (p.selected = true) : ''));
					}}>Select All</Button
				>
				<Button
					onclick={() => {
						pageMode = 'display';
						pages.forEach((p) => (p.selected ? (p.selected = false) : ''));
					}}>Cancel Selection</Button
				>
			</div>
		{/if}
	</div>
</div>

<div class=" m-4 grid grid-cols-[repeat(auto-fill,_minmax(150px,_1fr))] gap-x-6 gap-y-4">
	<!-- <div class="grid grid-cols-2 sm:grid-cols-2 md:grid-cols-5 lg:grid-cols-5 gap-2 justify-items-center"> -->
	{#each pages as page}
	
		<PageThumbnail
			id={page.id.toString()}
			ratio={720 / 1080}
			src={PageTumbnailURL(page.id)}
			alt="page {page.id} thumbnail"
			mode={pageMode}
			href="/page/{page.id}/"
			bind:checked={page.selected}
		></PageThumbnail>
	{/each}
</div>
