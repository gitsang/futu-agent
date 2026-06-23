<script lang="ts">
	import { cn } from '$lib/utils';

	interface Props {
		size?: 'sm' | 'md' | 'lg';
		class?: string;
	}

	let { size = 'md', class: className = '' }: Props = $props();

	const sizes = {
		sm: 'h-4 w-4',
		md: 'h-8 w-8',
		lg: 'h-12 w-12'
	};
</script>

<div
	class={cn(
		'animate-spin rounded-full border-2 border-border border-t-accent',
		sizes[size],
		className
	)}
	role="status"
>
	<span class="sr-only">Loading...</span>
</div>
