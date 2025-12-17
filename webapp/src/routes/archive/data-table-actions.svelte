<script lang="ts">
	import EllipsisIcon from '@lucide/svelte/icons/ellipsis';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { GenrateChapterFromArchive, GenratePagesFromArchive } from '$lib/data/api';
	import { toast } from 'svelte-sonner';
	import { goto, pushState } from '$app/navigation';
	import type { Archive } from '$lib/data/model';

	let { archive }: { archive: Archive } = $props();
</script>

<DropdownMenu.Root>
	<DropdownMenu.Trigger>
		{#snippet child({ props })}
			<Button {...props} variant="ghost" size="icon" class="relative size-8 p-0">
				<span class="sr-only">Open menu</span>
				<EllipsisIcon />
			</Button>
		{/snippet}
	</DropdownMenu.Trigger>
	<DropdownMenu.Content>
		<DropdownMenu.Group>
			<DropdownMenu.Label>View</DropdownMenu.Label>
			<a href="/archive/{archive.id}/pages">
				<DropdownMenu.Item
					onclick={() =>
						goto(`/archive/${archive.id}/pages`, {
							state: { archive }
						})}
				>
					View Pages
				</DropdownMenu.Item>
			</a>
		</DropdownMenu.Group>

		<DropdownMenu.Group>
			<DropdownMenu.Label>Generate</DropdownMenu.Label>

			<DropdownMenu.Item
				onclick={() => {
					let prom = GenratePagesFromArchive(archive.id);
					toast.promise(prom, {
						loading: 'Generating pages...',
						success: (data) => `Succesfully generated ${data.length} pages`,
						error: (err) => `Failed to generate pages: ${err}`,
						richColors: true,
						position: 'top-left',
						closeButton: true
					});
				}}
			>
				Generate Pages
			</DropdownMenu.Item>
			<DropdownMenu.Item
				onclick={() => {
					let prom = GenrateChapterFromArchive(archive.id);
					toast.promise(prom, {
						loading: 'Generating chapter...',
						success: (data) => `Succesfully generated chapter ${data.name}`,
						error: (err) => `Failed to generate chapter: ${err}`,
						richColors: true,
						position: 'top-left',
						closeButton: true
					});
				}}
			>
				Generate Chapter
			</DropdownMenu.Item>
		</DropdownMenu.Group>
	</DropdownMenu.Content>
</DropdownMenu.Root>
