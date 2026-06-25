<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	interface Props {
		variant?: 'default' | 'secondary' | 'destructive' | 'outline' | 'ghost' | 'link';
		size?: 'default' | 'sm' | 'lg' | 'icon';
		class?: string;
		disabled?: boolean;
		onclick?: (e: MouseEvent) => void;
		children: Snippet;
	}

	let {
		variant = 'default',
		size = 'default',
		class: className = '',
		disabled = false,
		onclick,
		children
	}: Props = $props();

	const variants = {
		default: 'bg-primary text-primary-foreground shadow-xs hover:bg-primary/90',
		secondary: 'bg-secondary text-secondary-foreground shadow-xs hover:bg-secondary/80',
		destructive: 'bg-destructive text-white shadow-xs hover:bg-destructive/90 focus-visible:ring-destructive/20 dark:focus-visible:ring-destructive/40',
		outline: 'border border-input bg-background shadow-xs hover:bg-accent hover:text-accent-foreground',
		ghost: 'hover:bg-accent hover:text-accent-foreground',
		link: 'text-primary underline-offset-4 hover:underline'
	};

	const sizes = {
		default: 'h-9 px-4 py-2 has-[>svg]:px-3',
		sm: 'h-8 rounded-md gap-1.5 px-3 has-[>svg]:px-2.5',
		lg: 'h-10 rounded-md px-6 has-[>svg]:px-4',
		icon: 'size-9'
	};
</script>

<button
	class={cn(
		'inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0 cursor-pointer',
		variants[variant],
		sizes[size],
		className
	)}
	{disabled}
	{onclick}
>
	{@render children()}
</button>
