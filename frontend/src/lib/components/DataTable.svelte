<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	interface Column {
		key: string;
		label: string;
		class?: string;
		align?: 'left' | 'center' | 'right';
	}

	interface Props {
		columns: Column[];
		data: Record<string, unknown>[];
		loading?: boolean;
		emptyMessage?: string;
		rowSnippet?: Snippet<[{ row: Record<string, unknown>; column: Column }]>;
	}

	let {
		columns,
		data,
		loading = false,
		emptyMessage = '暂无数据',
		rowSnippet
	}: Props = $props();
</script>

<div class="overflow-x-auto">
	<table class="w-full">
		<thead>
			<tr class="border-b border-border">
				{#each columns as column}
					<th
						class={cn(
							'px-4 py-3 text-xs font-medium text-muted-foreground uppercase tracking-wider',
							column.align === 'center' && 'text-center',
							column.align === 'right' && 'text-right',
							column.class
						)}
					>
						{column.label}
					</th>
				{/each}
			</tr>
		</thead>
		<tbody>
			{#if loading}
				{#each Array(3) as _}
					<tr class="border-b border-border-subtle">
						{#each columns as _}
							<td class="px-4 py-3">
								<div class="h-4 w-20 animate-pulse rounded bg-muted"></div>
							</td>
						{/each}
					</tr>
				{/each}
			{:else if data.length === 0}
				<tr>
					<td colspan={columns.length} class="px-4 py-8 text-center text-muted-foreground">
						{emptyMessage}
					</td>
				</tr>
			{:else}
				{#each data as row}
					<tr class="border-b border-border-subtle transition-colors hover:bg-muted/50">
						{#each columns as column}
							<td
								class={cn(
									'px-4 py-3 text-sm',
									column.align === 'center' && 'text-center',
									column.align === 'right' && 'text-right',
									column.class
								)}
							>
								{#if rowSnippet}
									{@render rowSnippet({ row, column })}
								{:else}
									{row[column.key] ?? '-'}
								{/if}
							</td>
						{/each}
					</tr>
				{/each}
			{/if}
		</tbody>
	</table>
</div>
