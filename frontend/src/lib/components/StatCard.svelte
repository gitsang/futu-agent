<script lang="ts">
	import Card from './Card.svelte';
	import { cn } from '$lib/utils';

	interface Props {
		label: string;
		value: string;
		subtitle?: string;
		variant?: 'default' | 'profit' | 'loss';
		loading?: boolean;
	}

	let { label, value, subtitle, variant = 'default', loading = false }: Props = $props();
</script>

<Card>
	<div class="flex flex-col gap-1">
		<span class="text-sm text-muted-foreground">{label}</span>
		{#if loading}
			<div class="h-8 w-24 animate-pulse rounded bg-muted"></div>
		{:else}
			<span
				class={cn(
					'text-2xl font-semibold tracking-tight font-mono',
					variant === 'profit' && 'text-profit',
					variant === 'loss' && 'text-loss'
				)}
			>
				{value}
			</span>
		{/if}
		{#if subtitle}
			<span class="text-xs text-muted-foreground">{subtitle}</span>
		{/if}
	</div>
</Card>
