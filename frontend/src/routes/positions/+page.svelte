<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import type { Position } from '$lib/types';
	import { formatCurrency, formatNumber, cn } from '$lib/utils';
	import { Card, StatCard, Badge, LoadingSpinner, Button } from '$lib/components';
	import { selectedMarket } from '$lib/stores';

	let positions = $state<Position[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let sortBy = $state<'stock_code' | 'unrealized_pnl' | 'market_value'>('stock_code');
	let sortOrder = $state<'asc' | 'desc'>('asc');

	onMount(async () => {
		try {
			positions = await api.getPositions();
		} catch (e) {
			error = e instanceof Error ? e.message : '加载失败';
		} finally {
			loading = false;
		}
	});

	let filteredPositions = $derived(
		$selectedMarket === 'ALL'
			? positions
			: positions.filter(p => p.market === $selectedMarket)
	);

	let sortedPositions = $derived([...filteredPositions].sort((a, b) => {
		let comparison = 0;
		switch (sortBy) {
			case 'stock_code':
				comparison = a.code.localeCompare(b.code);
				break;
			case 'unrealized_pnl':
				comparison = a.unrealized_pnl - b.unrealized_pnl;
				break;
			case 'market_value':
				comparison = (a.current_price * a.quantity) - (b.current_price * b.quantity);
				break;
		}
		return sortOrder === 'asc' ? comparison : -comparison;
	}));

	let totalPnl = $derived(filteredPositions.reduce((sum, p) => sum + p.unrealized_pnl, 0));
	let totalMarketValue = $derived(filteredPositions.reduce((sum, p) => sum + (p.current_price * p.quantity), 0));
	let totalCost = $derived(filteredPositions.reduce((sum, p) => sum + (p.avg_cost * p.quantity), 0));
	let totalPnlPercent = $derived(totalCost > 0 ? (totalPnl / totalCost) * 100 : 0);

	function toggleSort(column: typeof sortBy) {
		if (sortBy === column) {
			sortOrder = sortOrder === 'asc' ? 'desc' : 'asc';
		} else {
			sortBy = column;
			sortOrder = 'asc';
		}
	}

	function getPnlColor(pnl: number): string {
		if (pnl > 0) return 'text-profit';
		if (pnl < 0) return 'text-loss';
		return 'text-text-secondary';
	}

	function getPnlBg(pnl: number): string {
		if (pnl > 0) return 'bg-profit-bg';
		if (pnl < 0) return 'bg-loss-bg';
		return 'bg-surface-elevated';
	}
</script>

<div class="space-y-6 animate-fade-in">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold text-text-primary">持仓管理</h1>
			<p class="text-sm text-text-secondary">当前股票持仓与盈亏分析</p>
		</div>
		<Button variant="secondary" onclick={() => window.location.reload()}>
			<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
			</svg>
			刷新
		</Button>
	</div>

	{#if error}
		<Card class="border-loss/50 bg-loss/5">
			<div class="flex items-center gap-3 text-loss">
				<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
				</svg>
				<span class="text-sm">{error}</span>
			</div>
		</Card>
	{/if}

	<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
		<StatCard
			label="持仓市值"
			value={formatCurrency(totalMarketValue)}
			{loading}
		/>
		<StatCard
			label="持仓成本"
			value={formatCurrency(totalCost)}
			{loading}
		/>
		<StatCard
			label="总盈亏"
			value={`${totalPnl >= 0 ? '+' : ''}${formatCurrency(totalPnl)}`}
			subtitle={formatNumber(totalPnlPercent) + '%'}
			variant={totalPnl >= 0 ? 'profit' : 'loss'}
			{loading}
		/>
		<StatCard
			label="持仓数量"
			value={positions.length.toString()}
			subtitle="只股票"
			{loading}
		/>
	</div>

	<Card padding={false}>
		<div class="overflow-x-auto">
			<table class="w-full">
				<thead>
					<tr class="border-b border-border">
						<th
							class="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider hover:text-text-secondary cursor-pointer"
							onclick={() => toggleSort('stock_code')}
							onkeydown={(e) => e.key === 'Enter' && toggleSort('stock_code')}
							tabindex="0"
							role="columnheader"
							aria-sort={sortBy === 'stock_code' ? (sortOrder === 'asc' ? 'ascending' : 'descending') : 'none'}
						>
							股票代码 {sortBy === 'stock_code' ? (sortOrder === 'asc' ? '↑' : '↓') : ''}
						</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">市场</th>
						<th class="px-4 py-3 text-right text-xs font-medium text-text-muted uppercase tracking-wider">持仓数量</th>
						<th class="px-4 py-3 text-right text-xs font-medium text-text-muted uppercase tracking-wider">成本价</th>
						<th class="px-4 py-3 text-right text-xs font-medium text-text-muted uppercase tracking-wider">现价</th>
						<th
							class="px-4 py-3 text-right text-xs font-medium text-text-muted uppercase tracking-wider hover:text-text-secondary cursor-pointer"
							onclick={() => toggleSort('unrealized_pnl')}
							onkeydown={(e) => e.key === 'Enter' && toggleSort('unrealized_pnl')}
							tabindex="0"
							role="columnheader"
							aria-sort={sortBy === 'unrealized_pnl' ? (sortOrder === 'asc' ? 'ascending' : 'descending') : 'none'}
						>
							盈亏 {sortBy === 'unrealized_pnl' ? (sortOrder === 'asc' ? '↑' : '↓') : ''}
						</th>
						<th class="px-4 py-3 text-right text-xs font-medium text-text-muted uppercase tracking-wider">盈亏率</th>
					</tr>
				</thead>
				<tbody>
					{#if loading}
						{#each Array(3) as _}
							<tr class="border-b border-border-subtle">
								{#each Array(7) as _}
									<td class="px-4 py-3">
										<div class="h-4 w-16 animate-pulse rounded bg-surface-elevated"></div>
									</td>
								{/each}
							</tr>
						{/each}
					{:else if sortedPositions.length === 0}
						<tr>
							<td colspan="7" class="px-4 py-12 text-center text-text-muted">
								暂无持仓数据
							</td>
						</tr>
					{:else}
					{#each sortedPositions as pos}
						{@const pnlPercent = pos.avg_cost > 0 ? ((pos.current_price - pos.avg_cost) / pos.avg_cost) * 100 : 0}
						{@const marketValue = pos.current_price * pos.quantity}
						<tr class="border-b border-border-subtle transition-colors hover:bg-surface-hover">
							<td class="px-4 py-3">
								<div class="font-medium text-text-primary">{pos.code}</div>
							</td>
								<td class="px-4 py-3">
									<Badge variant="default">{pos.market}</Badge>
								</td>
								<td class="px-4 py-3 text-right font-mono text-sm text-text-primary">
									{formatNumber(pos.quantity, 0)}
								</td>
								<td class="px-4 py-3 text-right font-mono text-sm text-text-secondary">
									{formatCurrency(pos.avg_cost)}
								</td>
								<td class="px-4 py-3 text-right font-mono text-sm text-text-primary">
									{formatCurrency(pos.current_price)}
								</td>
								<td class="px-4 py-3 text-right">
									<span class={cn('font-mono text-sm font-medium', getPnlColor(pos.unrealized_pnl))}>
										{pos.unrealized_pnl >= 0 ? '+' : ''}{formatCurrency(pos.unrealized_pnl)}
									</span>
								</td>
								<td class="px-4 py-3 text-right">
									<span class={cn('inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium', getPnlBg(pos.unrealized_pnl), getPnlColor(pos.unrealized_pnl))}>
										{pnlPercent >= 0 ? '+' : ''}{formatNumber(pnlPercent)}%
									</span>
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</Card>
</div>
