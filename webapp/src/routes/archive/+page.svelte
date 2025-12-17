<script lang="ts">
	import type { Archive } from '$lib/data/model';
	import DataTable from './data-table.svelte';
	import { columns } from './columns.js';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { invalidateAll } from '$app/navigation';
	import { Ellipsis } from "@lucide/svelte";
	import Button from '$lib/components/ui/button/button.svelte';
	import { ScanContentDir } from '$lib/data/api';

	let { data } = $props();


	let archives: Archive[] = $state([]);
	$effect(()=>{
		toast.promise(data.archives, {
			loading: 'Fetching archives...',

			success: (data) => {
				archives = data;
				return data.length + ' archives fetched';
			},
			error: (err) => 'Error encountered ' + err,
			position: 'top-left',
			closeButton: true,
			richColors: true
		});
	})
</script>

<div class="flex justify-end p-3">
	<Button onclick={()=>ScanContentDir(globalThis.fetch)}>Scan Content Directory</Button>
</div>

<div class="mx-4 my-4">
	<DataTable data={archives} {columns} />
</div>

