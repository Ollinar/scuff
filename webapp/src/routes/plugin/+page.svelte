<script lang="ts">
	import * as Accordion from '$lib/components/ui/accordion/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Toggle } from '$lib/components/ui/toggle/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { PostPlugin, PostPluginEndpoint, UpdatePluginConfig } from '$lib/data/api.js';
	import { isHttpError } from '@sveltejs/kit';

	import { toast } from 'svelte-sonner';

	let { data } = $props();
	let plugins = $state(data.plugins);

	let pluginFile: FileList | undefined = $state();
</script>

<div class="flex justify-end p-4">
	<Dialog.Root>
		<Dialog.Trigger>Add Plugin</Dialog.Trigger>
		<Dialog.Content>
			<Dialog.Header>
				<Dialog.Title>Select Plugin file</Dialog.Title>
				<Dialog.Description>
					<Input type="file" name="file" bind:files={pluginFile} required />
					<Button
						type="submit"
						onclick={() => {
							if (!pluginFile || pluginFile.length != 1) {
								return;
							}
							const fd = new FormData();
							fd.set('file', pluginFile[0]);
							toast.promise(PostPlugin(globalThis.fetch, fd), {
								loading: 'Uploading plugin...',
								richColors: true,
								position: 'top-left',
								closeButton: true,
								error: (err) => {
									return `Failed to upload plugin: ${isHttpError(err) ? err.toString() : err}`;
								},
								success: () => {
									return `Plugin uploaded succesfully!`;
								}
							});
							console.log(fd.get('file'));
						}}>Submit</Button
					>
				</Dialog.Description>
			</Dialog.Header>
		</Dialog.Content>
	</Dialog.Root>
</div>

<Accordion.Root type="multiple" class="mx-auto w-[50%]">
	{#each plugins as plug, index}
		<Accordion.Item value="item-1">
			<Accordion.Trigger>{plug.name} Version: {plug.version}</Accordion.Trigger>
			<Accordion.Content class="max-h-[50%] overflow-auto rounded-md bg-card p-5">
				<p>{plug.description}</p>
				<p>Target Type: {plug.targetType}</p>
				<div>
					<Label for={`${index}-delay-input`}>Delay</Label>
					<Input
						id={`${index}-delay-input`}
						placeholder="Delay"
						type="number"
						bind:value={plug.delay}
					/>
					<Toggle aria-label="Toggle autorun" bind:pressed={plug.autoRun}>Auto Run</Toggle>
					{#each plug.config as config, confIdx}
						<Label class="m-2" for={`${index}-conf-input-${confIdx}`}>{config.name}</Label>
						<Input
							id={`${index}-conf-input-${confIdx}`}
							placeholder={config.name}
							bind:value={config.value}
						/>
					{/each}

					<div class="flex justify-end p-3">
						<Button
							onclick={() => {
								toast.promise(UpdatePluginConfig(globalThis.fetch, plug), {
									loading: 'Updating plugin config...',
									richColors: true,
									position: 'top-left',
									closeButton: true,
									error: (err) => {
										return `Failed to update plugin config: ${isHttpError(err) ? err.toString() : err}`;
									},
									success: () => {
										return `Plugin config updated succesfully!`;
									}
								});
							}}>Save</Button
						>
					</div>
				</div>
			</Accordion.Content>
		</Accordion.Item>
	{/each}
</Accordion.Root>
