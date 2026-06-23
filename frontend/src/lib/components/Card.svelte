<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	interface Props {
		class?: string;
		padding?: boolean;
		children: Snippet;
	}

	let { class: className = '', padding = true, children }: Props = $props();
</script>

<div
	class={cn(
		'rounded-xl border border-border bg-surface overflow-hidden',
		padding && 'p-6',
		className
	)}
>
	{@render children()}
</div>
