<script lang="ts">
	import { AspectRatio } from '$lib/components/ui/aspect-ratio/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { cn, type WithElementRef } from '$lib/utils';
	import type { HTMLAttributes } from 'svelte/elements';

	type PageThumbnailProp = {
		src: string;
		alt: string;
		ratio?: number;
		checked?: boolean;
		href?: string;
		mode: 'select' | 'navigate' | 'display';
	};

	let {
		ref = $bindable(null),
		id,
		src,
		alt,
		class: className,
		ratio,
		mode,
		checked = $bindable(false),
		href = '',
		children,
		...restProps
	}: PageThumbnailProp & WithElementRef<HTMLAttributes<HTMLDivElement>> = $props();
</script>

<div bind:this={ref} class={className} {...restProps}>
	{#if mode === 'select'}
		<Label>
			<div class="relative max-w-[250px] min-w-[150px] rounded-md bg-muted p-2 active:bg-ring">
				<div class="rounded-md bg-muted">
					<AspectRatio {ratio} class="self-center justify-self-center">
						<img {src} {alt} class="object-cover" />
					</AspectRatio>
					<Checkbox class="absolute top-2 right-2" bind:checked></Checkbox>
				</div>
			</div>
		</Label>
	{:else if mode === 'navigate'}
		<a {href}>
			<div class="relative max-w-[250px] min-w-[150px] rounded-md bg-muted p-2 active:bg-ring">
				<div class="rounded-md bg-muted">
					<AspectRatio {ratio} class="self-center justify-self-center">
						<img {src} {alt} class="object-cover" />
					</AspectRatio>
				</div>
			</div>
		</a>
	{:else}
		<div class="relative max-w-[250px] min-w-[150px] rounded-md bg-muted p-2 ">
			<div class="rounded-md bg-muted">
				<AspectRatio {ratio} class="self-center justify-self-center">
					<img {src} {alt} class="object-cover" />
				</AspectRatio>
			</div>
		</div>
	{/if}
</div>
