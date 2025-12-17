<script lang="ts">
	import * as Accordion from '$lib/components/ui/accordion/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Toggle } from '$lib/components/ui/toggle/index.js';
	import { toast } from 'svelte-sonner';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { isHttpError } from '@sveltejs/kit';
	import { ExecutePlugin } from '$lib/data/api.js';
	let { data } = $props();
	let plugins = $state(data.plugins);
</script>

<Accordion.Root type="multiple" class="mx-auto w-[50%]">
	{#each plugins as plug, index}
		<Accordion.Item value="item-1">
			<Accordion.Trigger>{plug.name} Version: {plug.version}</Accordion.Trigger>
			<Accordion.Content class="max-h-[50%] overflow-auto rounded-md bg-card p-5">
				<p>{plug.description}</p>
				<p>Target Type: {plug.targetType}</p>
				<div>
					{#each plug.param as param, paramIndx}
						<Label class="m-2" for={`${index}-param-input-${paramIndx}`}>{param.name}</Label>
						<Input
							id={`${index}-conf-input-${paramIndx}`}
							placeholder={param.name}
							bind:value={param.value}
						/>
					{/each}

					<div class="flex justify-end p-3">
						<Button
							onclick={() => {
								toast.promise(ExecutePlugin(globalThis.fetch,data.chapter.id, plug), {
									loading: 'Executing plugin...',
									richColors: true,
									position: 'top-left',
									closeButton: true,
									error: (err) => {
										return `Failed to execute plugin: ${isHttpError(err) ? err.toString() : err}`;
									},
									success: () => {
										return `Plugin executed succesfully!`;
									}
								});
							}}>Execute</Button
						>
					</div>
				</div>
			</Accordion.Content>
		</Accordion.Item>
	{/each}
</Accordion.Root>
