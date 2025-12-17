<script lang="ts">
	import { PageThumbnail, Root } from '$lib/components/ui/page-thumbnail';
	import { PageTumbnailURL } from '$lib/data/api';
	import * as Card from '$lib/components/ui/card/index.js';
	import { AspectRatio } from '$lib/components/ui/aspect-ratio/index.js';
	import Badge from '$lib/components/ui/badge/badge.svelte';
	import { Input } from '$lib/components/ui/input/index.js';
	import * as Pagination from '$lib/components/ui/pagination/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { navigating, page } from '$app/state';
	import { afterNavigate, goto, invalidateAll, replaceState } from '$app/navigation';
	import { onMount } from 'svelte';
	import { Search } from '@lucide/svelte';

	let { data } = $props();

	$effect(() => {
		const sp = new URLSearchParams(page.url.searchParams);
		sp.set('ps', data.perPage.toString());
		// artificial delay, so replace state doesn't error because the router isnt ready
		setTimeout(() => {
			replaceState(`?${sp.toString()}`, {});
		}, 0);
	});

	let sort = $derived.by(() => {
		if (sortSelect == 'tag') {
			return page.url.searchParams.get('sort') ?? '';
		}
		return '';
	});

	let sortNamespace = $state('');

	let sortSelect = $derived.by(() => {
		switch (page.url.searchParams.get('sort') ?? '') {
			case '':
			case 'id':
				return 'id';
			case 'name':
				return 'name';
			default:
				return 'tag';
		}
	});
	let sortDesc = $derived.by(() => {
		if (page.url.searchParams.get('sortDesc') == 'true') {
			return 'true';
		}
		return 'false';
	});
	let pageN = $derived.by(() => {
		const pg = parseInt(page.url.searchParams.get('p') ?? '');
		return !isNaN(pg) ? pg : 1;
	});
</script>

<div class="px-3 mx-auto flex md:w-[65%] flex-col-reverse py-3 gap-y-2 md:flex-row">
	<Select.Root type="single" bind:value={sortSelect}>
		<Select.Trigger class=""
			>Sort by
			{#if sortSelect == 'id'}
				ID
			{:else if sortSelect == 'name'}
				Name
			{:else}
				{sortNamespace}
			{/if}
		</Select.Trigger>
		<Select.Content>
			<Select.Item value="id">ID</Select.Item>
			<Select.Item value="name">Name</Select.Item>
			<Select.Item value="tag">Tag</Select.Item>
			<Input bind:value={sortNamespace} placeholder="namespace" />
		</Select.Content>
	</Select.Root>
	<Select.Root type="single" bind:value={sortDesc}>
		<Select.Trigger class="">{sortDesc == 'true' ? 'Descending' : 'Ascending'}</Select.Trigger>
		<Select.Content>
			<Select.Item value="false">Ascending</Select.Item>
			<Select.Item value="true">Descending</Select.Item>
		</Select.Content>
	</Select.Root>
	<form action="" method="get" class="flex grow">
		<input name="sort" hidden value={`${sortSelect == 'tag' ? sortNamespace : sortSelect}`} />
		<input name="sortDesc" hidden value={`${sortDesc == 'true'}`} />
		<Input name="q" placeholder="Search" value={page.url.searchParams.get('q') ?? ''} />
		<button><Search /></button>
	</form>
</div>

<p class=" text-center">
	Showing {data.perPage * (pageN - 1) + 1} - {data.perPage * (pageN - 1) + data.chapters.length}.
	Total of {data.totalChapter}
</p>

<div class="mx-3 my-2">
	<Select.Root
		type="single"
		bind:value={
			() => `${data.perPage}`,
			(v) => {
				const url = page.url;
				url.searchParams.set('ps', v);
				goto(url, { replaceState: true, invalidateAll: true });
			}
		}
	>
		<Select.Trigger class="">{data.perPage} Per Page</Select.Trigger>
		<Select.Content>
			{#each [10, 20, 30, 40, 50, 100] as pageSize}
				<Select.Item value={pageSize.toString()}>{pageSize}</Select.Item>
			{/each}
		</Select.Content>
	</Select.Root>
</div>

<div class="mx-3 grid grid-cols-1 gap-x-6 gap-y-4 border-2 p-2 sm:grid-cols-2">
	{#await data.chapters then chapters}
		{#each chapters as chapter}
			<a href={`/chapter/${chapter.id}`}>
				<div class="flex flex-col rounded-md border-2 bg-card p-3 sm:flex-row">
					<div class="mx-auto max-h-[300px] max-w-[200px] min-w-[200px] bg-border">
						<AspectRatio ratio={720 / 1080} class="self-center justify-self-center">
							<img
								src={`${PageTumbnailURL(chapter.coverPageId)}`}
								alt={`chapter ${chapter.id} cover`}
							/>
						</AspectRatio>
					</div>

					<div class="flex max-h-[300px] flex-grow flex-col px-3">
						<div class="mb-2 font-semibold">
							{chapter.name}
						</div>
						<div class="flex-1 overflow-y-auto pt-3">
							{#each chapter.tags as tag}
								<Badge>{tag.namespace}:{tag.label}</Badge>
							{/each}
						</div>
					</div>
				</div>
			</a>
		{/each}
	{/await}
</div>

<div>
	<Pagination.Root bind:page={pageN} count={data.totalChapter} perPage={data.perPage}>
		{#snippet children({ pages, currentPage })}
			<Pagination.Content>
				<Pagination.Item>
					<Pagination.PrevButton
						onclick={() => {
							const url = page.url;
							url.searchParams.set('p', (pageN - 1).toString());
							goto(url, { invalidateAll: true });
						}}
					/>
				</Pagination.Item>
				{#each pages as pageItem (pageItem.key)}
					{#if pageItem.type === 'ellipsis'}
						<Pagination.Item>
							<Pagination.Ellipsis />
						</Pagination.Item>
					{:else}
						<Pagination.Item>
							<Pagination.Link
								onclick={() => {
									const url = page.url;
									url.searchParams.set('p', pageItem.value.toString());
									goto(url, {
										invalidateAll: true
									});
								}}
								page={pageItem}
								isActive={currentPage === pageItem.value}
							>
								{pageItem.value}
							</Pagination.Link>
						</Pagination.Item>
					{/if}
				{/each}
				<Pagination.Item>
					<Pagination.NextButton
						onclick={() => {
							const url = page.url;
							url.searchParams.set('p', (pageN + 1).toString());
							goto(url, { invalidateAll: true });
						}}
					/>
				</Pagination.Item>
			</Pagination.Content>
		{/snippet}
	</Pagination.Root>
</div>
