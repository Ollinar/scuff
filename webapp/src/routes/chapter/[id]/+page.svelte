<script lang="ts">
	import { AspectRatio } from '$lib/components/ui/aspect-ratio';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { PageThumbnail } from '$lib/components/ui/page-thumbnail/index.js';
	import { DeleteChapter, PageTumbnailURL, PatchChapter } from '$lib/data/api';
	import { Pencil } from '@lucide/svelte';
	import * as Form from '$lib/components/ui/form/index.js';
	import { superForm, defaults, setError } from 'sveltekit-superforms';
	import { schemasafe } from 'sveltekit-superforms/adapters';
	import { Input } from '$lib/components/ui/input/index.js';
	import type { JSONSchema } from 'sveltekit-superforms';
	import Textarea from '$lib/components/ui/textarea/textarea.svelte';
	import type { Chapter } from '$lib/data/model.js';
	import { toast } from 'svelte-sonner';
	import { flip } from 'svelte/animate';
	import { dndzone, setDebugMode, type DndEvent } from 'svelte-dnd-action';
	import * as ContextMenu from '$lib/components/ui/context-menu/index.js';
	import { goto, invalidate, invalidateAll } from '$app/navigation';
	import { isHttpError } from '@sveltejs/kit';
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	export const chapterUpdateSchema = {
		type: 'object',
		properties: {
			name: { type: 'string', minLength: 1 },
			description: { type: 'string' },
			coverPage:{type:'integer'},
			tags: {
				type: 'array',
				items: {
					type: 'object',
					properties: {
						namespace: { type: 'string', minLength: 1 },
						label: { type: 'string', minLength: 1 }
					},
					required: ['namespace', 'label']
				}
			}
		},
		required: ['name', 'tags','coverPage'],
		additionalProperties: false,
		$schema: 'http://json-schema.org/draft-07/schema#'
	} as const satisfies JSONSchema;

	let { data } = $props();
	// let chapter = $derived(data.chapter);
	let deleteChapterDialog = $state(false);
	let editMetadataOpen = $state(false);

	const adapter = schemasafe(chapterUpdateSchema);
	const form = superForm(
		defaults(
			{
				name: data.chapter.name,
				coverPage:data.chapter.coverPageId,
				description: data.chapter.description,
				tags: data.chapter.tags.map((t) => {
					return { ...t };
				})
			},
			adapter
		),
		{
			SPA: true,
			dataType: 'json',
			validators: schemasafe(chapterUpdateSchema),
			async onUpdate({ form, result, cancel }) {
				if (!form.valid) {
					return;
				}
				const chap: Partial<Chapter> = {
					name: form.data.name,
					coverPageId:form.data.coverPage
				};
				if (
					form.data.description != undefined &&
					form.data.description != data.chapter.description
				) {
					chap.description = form.data.description;
				}
				chap.tags = [];
				chap.tags.push(...form.data.tags);
				const res = PatchChapter(globalThis.fetch, data.chapter.id, chap);
				toast.promise(res, {
					loading: 'Updating chapter...',
					richColors: true,
					position: 'top-left',
					closeButton: true,
					error: (err) => {
						return `Failed to update chapter: ${isHttpError(err) ? err.toString() : err}`;
					},
					success: (newdata) => {
						editMetadataOpen = false;
						result.data = newdata;
						return `Chapter updated succesfully!`;
					}
				});
				try {
					await res;
				} catch (error) {
					cancel();
				}
			}
		}
	);

	const { form: formData, enhance } = form;

	let newTagInput = $state({
		namespace: '',
		label: ''
	});
	function addToTags(event: KeyboardEvent) {
		if (event.key !== 'Enter') {
			return;
		}
		event.preventDefault();
		if (newTagInput.namespace != '' && newTagInput.label != '') {
			$formData.tags.push({
				namespace: newTagInput.namespace,
				label: newTagInput.label
			});
			$formData.tags = $formData.tags;
			newTagInput.namespace = '';
			newTagInput.label = '';
			return;
		}
	}

	let editingPages = $state(false);
	const flipDurationMs = 300;

	let pages: { id: number; pageId: number }[] = $state([]);

	pages = data.chapter.pageIDs.map((p, i) => {
		return { id: i + 1, pageId: p };
	});
	setDebugMode(false);
	function handleDndConsider(e: CustomEvent<DndEvent<{ id: number; pageId: number }>>) {
		pages = e.detail.items;
	}
	function handleDndFinalize(e: CustomEvent<DndEvent<{ id: number; pageId: number }>>) {
		pages = e.detail.items;
	}
</script>

<div class=" m-4 border-2 bg-card p-2">
	<div class="relative flex flex-col sm:flex-row">
		<DropdownMenu.Root>
			<DropdownMenu.Trigger disabled={editMetadataOpen || editingPages}>
				{#snippet child({ props })}
					<Button {...props} variant="ghost" class="absolute top-2 right-2 z-2">
						<Pencil />
					</Button>
				{/snippet}
			</DropdownMenu.Trigger>
			<DropdownMenu.Content align="start">
				<DropdownMenu.Item onclick={() => (editMetadataOpen = true)}
					>Edit Metadata</DropdownMenu.Item
				>
				<DropdownMenu.Item onclick={() => (editingPages = true)}>Edit Pages</DropdownMenu.Item>
				<DropdownMenu.Item onclick={() => (deleteChapterDialog = true)}
					>Delete Chapter</DropdownMenu.Item
				>
				<DropdownMenu.Item
					onclick={() => goto(`/chapter/${data.chapter.id}/execPlugin`)}>Execute Plugin</DropdownMenu.Item
				>
			</DropdownMenu.Content>
		</DropdownMenu.Root>

		<div class="mx-auto max-h-[300px] max-w-[200px] min-w-[200px] bg-border">
			<AspectRatio ratio={720 / 1080} class="self-center justify-self-center">
				<img
					src={`${PageTumbnailURL(data.chapter.coverPageId)}`}
					alt={`chapter ${data.chapter.id} cover`}
				/>
			</AspectRatio>
		</div>
		<div class="flex max-h-[300px] flex-grow flex-col px-3">
			<div class="mb-2 font-semibold">
				{data.chapter.name}
			</div>
			<p>{data.chapter.description}</p>
			<div class="flex-1 overflow-y-auto pt-3">
				{#each data.chapter.tags as tag}
					<Badge>{tag.namespace}:{tag.label}</Badge>
				{/each}
			</div>
		</div>
	</div>
	{#if editingPages}
		<div class="mt-2 flex justify-end gap-2">
			<Button
				onclick={() => {
					const newChap = data.chapter;
					newChap.pageIDs = pages.map((p) => p.pageId);
					toast.promise(PatchChapter(globalThis.fetch, newChap.id, newChap), {
						loading: 'Updating pages...',
						richColors: true,
						position: 'top-left',
						closeButton: true,
						error: (err) => `Failed to update pages: ${err}`,
						success: () => {
							editingPages = false;
							invalidateAll();
							return `Chapter pages updated succesfully!`;
						}
					});
				}}>Save</Button
			>
			<Button
				onclick={() =>
					(pages = data.chapter.pageIDs.map((p, i) => {
						return { id: i + 1, pageId: p };
					}))}>Reset</Button
			>
			<Button onclick={() => (editingPages = false)}>Cancel</Button>
		</div>
	{/if}
</div>

<div class="mx-auto w-fit">
	{`Total Pages: ${editingPages ? pages.length : data.chapter.pageIDs.length}`}
</div>

{#if editingPages}
	<div
		class=" m-4 grid grid-cols-[repeat(auto-fill,_minmax(150px,_1fr))] gap-x-6 gap-y-4"
		use:dndzone={{ items: pages, flipDurationMs, delayTouchStart: true }}
		onconsider={handleDndConsider}
		onfinalize={handleDndFinalize}
		aria-label="pages"
	>
		{#each pages as page, idx (page.id)}
			<div class="relative" animate:flip={{ duration: flipDurationMs }}>
				<ContextMenu.Root>
					<ContextMenu.Trigger>
						<!-- doing the attach insead of just potting the index inside the bagge
						this is so it acutually gets updated after the badge blur 
						-->
						<Badge
							{@attach (el) => {
								el.innerText = (idx + 1).toString();
							}}
							class="absolute top-2 right-2 z-1"
							ondblclick={(event) => {
								event.currentTarget.contentEditable = 'true';
								event.currentTarget.inputMode = 'number';
							}}
							onkeydown={(event) => {
								if (event.key !== 'Enter') {
									return;
								}
								event.preventDefault();
								const inp = parseInt(event.currentTarget.innerText);
								if (isNaN(inp) || inp < 1 || inp > pages.length) {
									event.currentTarget.innerText = (idx + 1).toString();
									return;
								}

								const clone = [...pages];
								const [pg] = clone.splice(idx, 1);
								clone.splice(inp - 1, 0, pg);
								pages = clone;
								event.currentTarget.contentEditable = 'false';
							}}
						></Badge>
						<PageThumbnail
							id={page.pageId.toString()}
							ratio={720 / 1080}
							src={PageTumbnailURL(page.pageId)}
							alt="page {idx + 1} thumbnail"
							mode={'display'}
							href="/chapter/{data.chapter.id}/reader#{idx + 1}"
						></PageThumbnail>
					</ContextMenu.Trigger>
					<ContextMenu.Content>
						<ContextMenu.Item
							onclick={(event) => {
								pages = pages.filter((p) => p.id != page.id);
							}}>Remove</ContextMenu.Item
						>
					</ContextMenu.Content>
				</ContextMenu.Root>
			</div>
		{/each}
	</div>
{:else}
	<div class=" m-4 grid grid-cols-[repeat(auto-fill,_minmax(150px,_1fr))] gap-x-6 gap-y-4">
		{#each data.chapter.pageIDs as page, pageNum}
			<PageThumbnail
				id={page.toString()}
				ratio={720 / 1080}
				src={PageTumbnailURL(page)}
				alt="page {pageNum + 1} thumbnail"
				mode={'navigate'}
				href="/chapter/{data.chapter.id}/reader#{pageNum + 1}"
			></PageThumbnail>
		{/each}
	</div>
{/if}

<Dialog.Root bind:open={editMetadataOpen}>
	<Dialog.Content class="max-h-screen overflow-auto">
		<Dialog.Header>
			<Dialog.Title>Update Chapter Metadata</Dialog.Title>
			<Dialog.Description></Dialog.Description>
		</Dialog.Header>
		<form use:enhance>
			<Form.Field {form} name="name">
				<Form.Control>
					{#snippet children({ props })}
						<Form.Label>Name</Form.Label>
						<Input {...props} bind:value={$formData.name} />
					{/snippet}
				</Form.Control>
			</Form.Field>
			<Form.Field {form} name="coverPage">
				<Form.Control>
					{#snippet children({ props })}
						<Form.Label>Cover Page</Form.Label>
						<Input {...props} bind:value={$formData.coverPage} type='number' />
					{/snippet}
				</Form.Control>
			</Form.Field>
			<Form.Field {form} name="description">
				<Form.Control>
					{#snippet children({ props })}
						<Form.Label>Description</Form.Label>
						<Textarea {...props} bind:value={$formData.description}></Textarea>
					{/snippet}
				</Form.Control>
			</Form.Field>
			<Form.Fieldset {form} name="tags" class="my-2 max-h-[200px] overflow-auto">
				{#each $formData.tags as _, i}
					<Badge variant="outline">
						<Form.ElementField {form} name="tags[{i}].namespace">
							<Form.Control>
								{#snippet children({ props })}
									<Input
										{@attach (node) => {}}
										class="tag-namespace-input-{i} field-sizing-content h-auto min-w-4 px-0 py-0 "
										{...props}
										bind:value={$formData.tags[i].namespace}
										onkeydown={(event) => {
											if (
												event.key === 'Backspace' &&
												$formData.tags[i].namespace === '' &&
												$formData.tags[i].label === ''
											) {
												event.preventDefault();
												$formData.tags = $formData.tags.filter((_, idx) => idx !== i);
											}
										}}
									/>
								{/snippet}
							</Form.Control>
						</Form.ElementField>
						:
						<Form.ElementField {form} name="tags[{i}].label">
							<Form.Control>
								{#snippet children({ props })}
									<Input
										class="field-sizing-content h-auto min-w-4 px-0 py-0 "
										{...props}
										bind:value={$formData.tags[i].label}
										onkeydown={(event) => {
											if (event.key === 'Backspace' && $formData.tags[i].label === '') {
												event.preventDefault();
												event.currentTarget.parentElement?.parentElement
													?.querySelector('input')
													?.focus();
											}
										}}
									/>
								{/snippet}
							</Form.Control>
						</Form.ElementField>
					</Badge>
				{/each}
				<Form.FieldErrors />
			</Form.Fieldset>
			<div class="mt-3 flex">
				<Input
					placeholder="Namespace"
					class="field-sizing-content h-auto w-auto min-w-25 px-0 py-0"
					bind:value={newTagInput.namespace}
					onkeydown={addToTags}
				/>
				<Input
					placeholder="Label"
					class="field-sizing-content h-auto w-auto min-w-25 px-0 py-0"
					bind:value={newTagInput.label}
					onkeydown={addToTags}
				/>
			</div>
			<Dialog.Footer>
				<Form.Button>Save changes</Form.Button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>
<AlertDialog.Root bind:open={deleteChapterDialog}>
	<AlertDialog.Content>
		<form
			onsubmit={(e) => {
				e.preventDefault();
				deleteChapterDialog = false;
				toast.promise(DeleteChapter(globalThis.fetch, data.chapter.id), {
					loading: 'Updating chapter...',
					richColors: true,
					position: 'top-left',
					closeButton: true,
					error: (err) => {
						return `Failed to update chapter: ${isHttpError(err) ? err.toString() : err}`;
					},
					success: () => {
						goto('/chapter');
						return `Chapter deleted succesfully!`;
					}
				});
			}}
		>
			<AlertDialog.Header>
				<AlertDialog.Title>Delete Chapter</AlertDialog.Title>
			</AlertDialog.Header>
			<AlertDialog.Footer>
				<AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
				<AlertDialog.Action>Continue</AlertDialog.Action>
			</AlertDialog.Footer>
		</form>
	</AlertDialog.Content>
</AlertDialog.Root>
