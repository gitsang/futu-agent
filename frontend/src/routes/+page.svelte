<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import type { AccountFunds, Decision, Position } from '$lib/types';
	import { formatCurrency, formatDate, cn } from '$lib/utils';
	import { StatCard, Card, Badge, LoadingSpinner, Button } from '$lib/components';
	import { selectedMarket } from '$lib/stores';

	let allFunds = $state<AccountFunds[]>([]);
	let positions = $state<Position[]>([]);
	let decisions = $state<Decision[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	onMount(async () => {
		try {
			const [fundsData, positionsData, decisionsData] = await Promise.all([
				api.getAllFunds(),
				api.getPositions(),
				api.getDecisions()
			]);
			allFunds = fundsData;
			positions = positionsData;
			decisions = decisionsData.slice(0, 5);
		} catch (e) {
			error = e instanceof Error ? e.message : '加载失败';
		} finally {
			loading = false;
		}
	});

	let filteredFunds = $derived(
		$selectedMarket === 'ALL'
			? allFunds.reduce((acc, f) => ({
					market: 'ALL',
					currency: 'CNY',
					total_assets: acc.total_assets + f.total_assets,
					cash: acc.cash + f.cash,
					market_value: acc.market_value + f.market_value
				}), { market: 'ALL', currency: 'CNY', total_assets: 0, cash: 0, market_value: 0 })
			: allFunds.find(f => f.market === $selectedMarket) || { market: $selectedMarket, currency: 'CNY', total_assets: 0, cash: 0, market_value: 0 }
	);

	let filteredPositions = $derived(
		$selectedMarket === 'ALL'
			? positions
			: positions.filter(p => p.market === $selectedMarket)
	);

	let filteredDecisions = $derived(
		$selectedMarket === 'ALL'
			? decisions
			: decisions.filter(d => d.market === $selectedMarket)
	);

	let totalPnl = $derived(
		filteredPositions.reduce((sum, p) => sum + p.unrealized_pnl, 0)
	);

	function getPnlClass(pnl: number): string {
		if (pnl > 0) return 'text-profit';
		if (pnl < 0) return 'text-loss';
		return 'text-text-secondary';
	}

	function getCurrencySymbol(currency: string): string {
		switch (currency) {
			case 'CNY': return '¥';
			case 'HKD': return 'HK$';
			case 'USD': return '$';
			default: return '¥';
		}
	}
</script>

<div class="space-y-6 animate-fade-in">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold text-text-primary">仪表盘</h1>
			<p class="text-sm text-text-secondary">账户概览与实时监控</p>
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
			label="总资产"
			value={loading ? '-' : `${getCurrencySymbol(filteredFunds.currency)}${filteredFunds.total_assets.toLocaleString()}`}
			{loading}
		/>
		<StatCard
			label="可用资金"
			value={loading ? '-' : `${getCurrencySymbol(filteredFunds.currency)}${filteredFunds.cash.toLocaleString()}`}
			{loading}
		/>
		<StatCard
			label="持仓市值"
			value={loading ? '-' : `${getCurrencySymbol(filteredFunds.currency)}${filteredFunds.market_value.toLocaleString()}`}
			{loading}
		/>
		<StatCard
			label="持仓盈亏"
			value={loading ? '-' : `${totalPnl >= 0 ? '+' : ''}${getCurrencySymbol(filteredFunds.currency)}${totalPnl.toLocaleString()}`}
			variant={totalPnl > 0 ? 'profit' : totalPnl < 0 ? 'loss' : 'default'}
			{loading}
		/>
	</div>

	<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
		<Card>
			<div class="flex items-center justify-between mb-4">
				<h2 class="text-lg font-semibold text-text-primary">当前持仓</h2>
				<a href="/positions" class="text-sm text-accent hover:text-accent-hover transition-colors">
					查看全部 →
				</a>
			</div>

			{#if loading}
				<div class="flex justify-center py-8">
					<LoadingSpinner />
				</div>
			{:else if filteredPositions.length === 0}
				<div class="py-8 text-center text-text-muted">暂无持仓</div>
			{:else}
				<div class="space-y-3">
					{#each filteredPositions.slice(0, 5) as pos}
						<div class="flex items-center justify-between rounded-lg bg-surface-elevated p-3 transition-colors hover:bg-surface-hover">
							<div>
								<div class="font-medium text-text-primary">{pos.code}</div>
								<div class="text-xs text-text-muted">{pos.market} · {pos.name} · {pos.quantity}股</div>
							</div>
							<div class="text-right">
								<div class="font-mono text-sm text-text-primary">
									{getCurrencySymbol(pos.market === 'CN' ? 'CNY' : pos.market === 'HK' ? 'HKD' : 'USD')}{pos.current_price.toLocaleString()}
								</div>
								<div class={cn('font-mono text-xs', getPnlClass(pos.unrealized_pnl))}>
									{pos.unrealized_pnl >= 0 ? '+' : ''}{pos.unrealized_pnl.toLocaleString()}
								</div>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</Card>

		<Card>
			<div class="flex items-center justify-between mb-4">
				<h2 class="text-lg font-semibold text-text-primary">最近决策</h2>
				<a href="/decisions" class="text-sm text-accent hover:text-accent-hover transition-colors">
					查看全部 →
				</a>
			</div>

			{#if loading}
				<div class="flex justify-center py-8">
					<LoadingSpinner />
				</div>
			{:else if filteredDecisions.length === 0}
				<div class="py-8 text-center text-text-muted">暂无决策记录</div>
			{:else}
				<div class="space-y-3">
					{#each filteredDecisions as decision}
						<div class="rounded-lg bg-surface-elevated p-3 transition-colors hover:bg-surface-hover">
							<div class="flex items-center justify-between mb-1">
								<div class="flex items-center gap-2">
									<span class="font-medium text-text-primary">{decision.stock_code}</span>
									<span class="text-xs text-text-muted">{decision.market}</span>
									<Badge
										variant={decision.action?.toLowerCase() === 'buy' ? 'success' : decision.action?.toLowerCase() === 'sell' ? 'destructive' : 'default'}
									>
										{decision.action?.toLowerCase() === 'buy' ? '买入' : decision.action?.toLowerCase() === 'sell' ? '卖出' : '持有'}
									</Badge>
								</div>
								<Badge variant={decision.executed ? 'success' : 'warning'}>
									{decision.executed ? '已执行' : '待执行'}
								</Badge>
							</div>
							<div class="text-xs text-text-muted line-clamp-1">{decision.reason}</div>
							<div class="mt-1 text-xs text-text-muted">{formatDate(decision.created_at)}</div>
						</div>
					{/each}
				</div>
			{/if}
		</Card>
	</div>
</div>
