<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import type { AccountFunds, Decision, MarketOverview, Position, TradingStats } from '$lib/types';
	import { formatCurrency, formatDate, cn } from '$lib/utils';
	import { StatCard, Card, Badge, LoadingSpinner, Button } from '$lib/components';
	import { selectedMarket } from '$lib/stores';

	let funds = $state<AccountFunds | null>(null);
	let positions = $state<Position[]>([]);
	let decisions = $state<Decision[]>([]);
	let tradingStats = $state<TradingStats | null>(null);
	let marketOverview = $state<MarketOverview[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	async function loadData() {
		loading = true;
		try {
			const market = $selectedMarket === 'ALL' ? undefined : $selectedMarket;
			const [fundsData, positionsData, decisionsResult, statsData, overviewData] = await Promise.all([
				api.getFunds(market),
				api.getPositions(market),
				api.getDecisions(market, 1, 5),
				api.getTradingStats(market),
				api.getMarketOverview()
			]);
			funds = fundsData;
			positions = positionsData;
			decisions = decisionsResult.data || [];
			tradingStats = statsData;
			marketOverview = overviewData;
		} catch (e) {
			error = e instanceof Error ? e.message : '加载失败';
		} finally {
			loading = false;
		}
	}

	onMount(() => loadData());

	$effect(() => {
		if ($selectedMarket) {
			loadData();
		}
	});

	let totalPnl = $derived(
		positions.reduce((sum, p) => sum + p.unrealized_pnl, 0)
	);

	function getPnlClass(pnl: number): string {
		if (pnl > 0) return 'text-profit';
		if (pnl < 0) return 'text-loss';
		return 'text-muted-foreground';
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
			<h1 class="text-2xl font-semibold text-foreground">仪表盘</h1>
			<p class="text-sm text-muted-foreground">账户概览与实时监控</p>
		</div>
		<Button variant="secondary" onclick={() => window.location.reload()}>
			<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
			</svg>
			刷新
		</Button>
	</div>

	{#if error}
		<Card class="border-destructive/50 bg-destructive/5">
			<div class="flex items-center gap-3 text-destructive">
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
			value={loading || !funds ? '-' : `${getCurrencySymbol(funds.currency)}${funds.total_assets.toLocaleString()}`}
			{loading}
		/>
		<StatCard
			label="可用资金"
			value={loading || !funds ? '-' : `${getCurrencySymbol(funds.currency)}${funds.cash.toLocaleString()}`}
			{loading}
		/>
		<StatCard
			label="持仓市值"
			value={loading || !funds ? '-' : `${getCurrencySymbol(funds.currency)}${funds.market_value.toLocaleString()}`}
			{loading}
		/>
		<StatCard
			label="持仓盈亏"
			value={loading || !funds ? '-' : `${totalPnl >= 0 ? '+' : ''}${getCurrencySymbol(funds.currency)}${totalPnl.toLocaleString()}`}
			variant={totalPnl > 0 ? 'profit' : totalPnl < 0 ? 'loss' : 'default'}
			{loading}
		/>
	</div>

	{#if tradingStats}
		<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
			<Card>
				<div class="text-sm text-muted-foreground">总订单数</div>
				<div class="text-2xl font-semibold text-foreground">{tradingStats.total_orders}</div>
			</Card>
			<Card>
				<div class="text-sm text-muted-foreground">成交订单</div>
				<div class="text-2xl font-semibold text-profit">{tradingStats.filled_orders}</div>
			</Card>
			<Card>
				<div class="text-sm text-muted-foreground">成功率</div>
				<div class="text-2xl font-semibold text-primary">{tradingStats.win_rate.toFixed(1)}%</div>
			</Card>
			<Card>
				<div class="text-sm text-muted-foreground">总成交量</div>
				<div class="text-2xl font-semibold text-foreground">{tradingStats.total_volume.toLocaleString()}</div>
			</Card>
		</div>
	{/if}

	{#if marketOverview.length > 0}
		<Card>
			<h2 class="text-lg font-semibold text-foreground mb-4">市场概览</h2>
			<div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
				{#each marketOverview as overview}
					<div class="rounded-lg bg-muted p-4">
						<div class="flex items-center justify-between mb-2">
							<span class="font-medium text-foreground">{overview.market}</span>
							<Badge variant={overview.total_pnl >= 0 ? 'success' : 'destructive'}>
								{overview.total_pnl >= 0 ? '+' : ''}{overview.total_pnl.toLocaleString()}
							</Badge>
						</div>
						<div class="text-sm text-muted-foreground">
							持仓: {overview.stock_count} 只 · 今日交易: {overview.today_trades} 笔
						</div>
					</div>
				{/each}
			</div>
		</Card>
	{/if}

	<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
		<Card>
			<div class="flex items-center justify-between mb-4">
				<h2 class="text-lg font-semibold text-foreground">当前持仓</h2>
				<a href="/positions" class="text-sm text-primary hover:text-primary/80 transition-colors">
					查看全部 →
				</a>
			</div>

			{#if loading}
				<div class="flex justify-center py-8">
					<LoadingSpinner />
				</div>
			{:else if positions.length === 0}
				<div class="py-8 text-center text-muted-foreground">暂无持仓</div>
			{:else}
				<div class="space-y-3">
					{#each positions.slice(0, 5) as pos}
						<div class="flex items-center justify-between rounded-lg bg-muted p-3 transition-colors hover:bg-muted/80">
							<div>
								<div class="font-medium text-foreground">{pos.code}</div>
								<div class="text-xs text-muted-foreground">{pos.market} · {pos.name} · {pos.quantity}股</div>
							</div>
							<div class="text-right">
								<div class="font-mono text-sm text-foreground">
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
				<h2 class="text-lg font-semibold text-foreground">最近决策</h2>
				<a href="/decisions" class="text-sm text-primary hover:text-primary/80 transition-colors">
					查看全部 →
				</a>
			</div>

			{#if loading}
				<div class="flex justify-center py-8">
					<LoadingSpinner />
				</div>
			{:else if decisions.length === 0}
				<div class="py-8 text-center text-muted-foreground">暂无决策记录</div>
			{:else}
				<div class="space-y-3">
					{#each decisions as decision}
						<div class="rounded-lg bg-muted p-3 transition-colors hover:bg-muted/80">
							<div class="flex items-center justify-between mb-1">
								<div class="flex items-center gap-2">
									<span class="font-medium text-foreground">{decision.stock_code}</span>
									<span class="text-xs text-muted-foreground">{decision.market}</span>
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
							<div class="text-xs text-muted-foreground line-clamp-1">{decision.reason}</div>
							<div class="mt-1 text-xs text-muted-foreground">{formatDate(decision.created_at)}</div>
						</div>
					{/each}
				</div>
			{/if}
		</Card>
	</div>
</div>
