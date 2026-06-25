<script lang="ts">
	import { cn } from '$lib/utils';
	import type { Snippet } from 'svelte';

	interface Props {
		variant?: 'default' | 'secondary' | 'destructive' | 'outline' | 'success' | 'warning' | 'info';
		class?: string;
		children: Snippet;
	}

	let { variant = 'default', class: className = '', children }: Props = $props();

	const variants = {
		default: 'border-transparent bg-primary text-primary-foreground shadow-xs',
		secondary: 'border-transparent bg-secondary text-secondary-foreground shadow-xs',
		destructive: 'border-transparent bg-destructive text-white shadow-xs',
		outline: 'text-foreground',
		success: 'border-transparent bg-profit text-white shadow-xs',
		warning: 'border-transparent bg-warning text-white shadow-xs',
		info: 'border-transparent bg-info text-white shadow-xs'
	};
</script>

<div
	class={cn(
		'inline-flex items-center rounded-md border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2',
		variants[variant],
		className
	)}
>
	{@render children()}
</div>
