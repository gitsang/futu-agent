<script lang="ts">
	import { cn } from '$lib/utils';

	interface Props {
		status: 'online' | 'offline' | 'warning' | 'unknown';
		label: string;
		class?: string;
	}

	let { status, label, class: className = '' }: Props = $props();

	const statusColors = {
		online: 'bg-profit',
		offline: 'bg-destructive',
		warning: 'bg-warning',
		unknown: 'bg-muted-foreground'
	};
</script>

<div class={cn('flex items-center gap-2', className)}>
	<div class={cn('h-2 w-2 rounded-full', statusColors[status])}></div>
	<span class="text-sm text-muted-foreground">{label}</span>
</div>
