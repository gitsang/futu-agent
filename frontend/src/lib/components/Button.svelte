<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	interface Props {
		variant?: 'default' | 'secondary' | 'destructive' | 'outline' | 'ghost';
		size?: 'sm' | 'md' | 'lg';
		class?: string;
		disabled?: boolean;
		onclick?: (e: MouseEvent) => void;
		children: Snippet;
	}

	let {
		variant = 'default',
		size = 'md',
		class: className = '',
		disabled = false,
		onclick,
		children
	}: Props = $props();

	const variants = {
		default: 'bg-accent text-white hover:bg-accent-hover',
		secondary: 'bg-surface-elevated text-text-primary hover:bg-surface-hover',
		destructive: 'bg-loss text-white hover:bg-loss/80',
		outline: 'border border-border text-text-primary hover:bg-surface-hover',
		ghost: 'text-text-secondary hover:text-text-primary hover:bg-surface-hover'
	};

	const sizes = {
		sm: 'px-3 py-1.5 text-xs',
		md: 'px-4 py-2 text-sm',
		lg: 'px-6 py-3 text-base'
	};
</script>

<button
	class={cn(
		'inline-flex items-center justify-center gap-2 rounded-lg font-medium transition-all duration-200 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed',
		variants[variant],
		sizes[size],
		className
	)}
	{disabled}
	{onclick}
>
	{@render children()}
</button>
