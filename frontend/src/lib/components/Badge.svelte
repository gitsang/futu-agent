<script lang="ts">
	import { cn } from '$lib/utils';

	interface Props {
		variant?: 'default' | 'success' | 'destructive' | 'warning' | 'info';
		class?: string;
		children: string;
	}

	let { variant = 'default', class: className = '', children }: Props = $props();

	const variants = {
		default: 'bg-surface-elevated text-text-secondary',
		success: 'bg-profit-bg text-profit',
		destructive: 'bg-loss-bg text-loss',
		warning: 'bg-warning/10 text-warning',
		info: 'bg-info/10 text-info'
	};
</script>

<span
	class={cn(
		'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium',
		variants[variant],
		className
	)}
>
	{children}
</span>
