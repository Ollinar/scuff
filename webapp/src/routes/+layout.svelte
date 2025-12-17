<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	import { ModeWatcher } from 'mode-watcher';
	import SunIcon from '@lucide/svelte/icons/sun';
	import MoonIcon from '@lucide/svelte/icons/moon';
	import { toggleMode } from 'mode-watcher';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Toaster } from '$lib/components/ui/sonner/index.js';

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<ModeWatcher />

<nav class="flex bg-secondary p-1 text-secondary-foreground">
	<ul class=" flex grow justify-center space-x-8 self-center">
		<li><a href="/archive" class="">Archives</a></li>
		<li><a href="/chapter" class="">Chapters</a></li>
		<li><a href="/plugin" class="">Plugins</a></li>
	</ul>
	<Button onclick={toggleMode} variant="outline" size="icon" class="self-end">
		<SunIcon
			class="h-[1.2rem] w-[1.2rem] scale-100 rotate-0 !transition-all dark:scale-0 dark:-rotate-90"
		/>
		<MoonIcon
			class="absolute h-[1.2rem] w-[1.2rem] scale-0 rotate-90 !transition-all dark:scale-100 dark:rotate-0"
		/>
		<span class="sr-only">Toggle theme</span>
	</Button>
</nav>
<Toaster />
{@render children?.()}
