<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import type { Order } from '$lib/types';
	import { cn, exportToCSV } from '$lib/utils';
	import { Card, Badge, LoadingSpinner, Button } from '$lib/components';
	import { selectedMarket } from '$lib/stores';

	let orders = $state<Order[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	async function loadOrders() {
		loading = true;
		try {
			const market = $selectedMarket === 'ALL' ? undefined : $selectedMarket;
			orders = await api.getOrders(market);
		} catch (e) {
			error = e instanceof Error ? e.message : '加载失败';
		} finally {
			loading = false;
		}
	}

	onMount(() => loadOrders());

	$effect(() => {
		if ($selectedMarket) {
			loadOrders();
		}
	});

	function handleExport() {
		const exportData = orders.map(o => ({
			'订单号': o.order_id,
			'股票代码': o.code,
			'股票名称': o.name,
			'市场': o.market,
			'方向': o.side === 'BUY' ? '买入' : '卖出',
			'委托价': o.price.toFixed(2),
			'委托量': o.qty,
			'成交量': o.fill_qty,
			'成交价': o.fill_price.toFixed(2),
			'状态': getStatusLabel(o.status),
			'下单时间': o.create_time
		}));
		exportToCSV(exportData, '订单数据');
	}

	function getStatusColor(status: string): 'success' | 'destructive' | 'warning' | 'default' {
		switch (status) {
			case 'FILLED_ALL':
			case 'FILLED_PART':
				return 'success';
			case 'CANCELLED_ALL':
			case 'CANCELLED_PART':
			case 'FAILED':
			case 'FILL_CANCELLED':
				return 'destructive';
			case 'WAITING_SUBMIT':
			case 'SUBMITTING':
			case 'SUBMITTED':
				return 'warning';
			default:
				return 'default';
		}
	}

	function getStatusLabel(status: string): string {
		switch (status) {
			case 'UNKNOWN': return '未知';
			case 'WAITING_SUBMIT': return '待提交';
			case 'SUBMITTING': return '提交中';
			case 'SUBMITTED': return '已提交';
			case 'FILLED_PART': return '部分成交';
			case 'FILLED_ALL': return '已成交';
			case 'CANCELLED_PART': return '部分撤单';
			case 'CANCELLED_ALL': return '已撤单';
			case 'FAILED': return '失败';
			case 'DISABLED': return '已失效';
			case 'DELETED': return '已删除';
			case 'FILL_CANCELLED': return '成交撤销';
			default: return status;
		}
	}

	function getSideLabel(side: string): string {
		return side === 'BUY' ? '买入' : '卖出';
	}

	function formatTime(time: string): string {
		if (!time) return '-';
		try {
			const date = new Date(time);
			return date.toLocaleString('zh-CN');
		} catch {
			return time;
		}
	}
</script>

<div class="space-y-6 animate-fade-in">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold text-foreground">订单管理</h1>
			<p class="text-sm text-muted-foreground">查看所有交易订单状态</p>
		</div>
		<div class="flex gap-2">
			<Button variant="secondary" onclick={handleExport}>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
				</svg>
				导出
			</Button>
			<Button variant="secondary" onclick={() => window.location.reload()}>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
				</svg>
				刷新
			</Button>
		</div>
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

	<Card padding={false}>
		<div class="overflow-x-auto">
			<table class="w-full">
				<thead>
					<tr class="border-b border-border">
						<th class="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">订单号</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">股票</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">市场</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">方向</th>
						<th class="px-4 py-3 text-right text-xs font-medium text-muted-foreground uppercase tracking-wider">委托价</th>
						<th class="px-4 py-3 text-right text-xs font-medium text-muted-foreground uppercase tracking-wider">委托量</th>
						<th class="px-4 py-3 text-right text-xs font-medium text-muted-foreground uppercase tracking-wider">成交量</th>
						<th class="px-4 py-3 text-right text-xs font-medium text-muted-foreground uppercase tracking-wider">成交价</th>
						<th class="px-4 py-3 text-center text-xs font-medium text-muted-foreground uppercase tracking-wider">状态</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">下单时间</th>
					</tr>
				</thead>
				<tbody>
					{#if loading}
						{#each Array(3) as _}
							<tr class="border-b border-border-subtle">
								{#each Array(10) as _}
									<td class="px-4 py-3">
										<div class="h-4 w-16 animate-pulse rounded bg-muted"></div>
									</td>
								{/each}
							</tr>
						{/each}
				{:else if orders.length === 0}
					<tr>
						<td colspan="10" class="px-4 py-12 text-center text-muted-foreground">
							暂无订单数据
						</td>
					</tr>
				{:else}
					{#each orders as order}
							<tr class="border-b border-border-subtle transition-colors hover:bg-muted/50">
								<td class="px-4 py-3">
									<div class="font-mono text-xs text-muted-foreground">{order.order_id?.slice(-8) || '-'}</div>
								</td>
								<td class="px-4 py-3">
									<div class="font-medium text-foreground">{order.code}</div>
									<div class="text-xs text-muted-foreground">{order.name}</div>
								</td>
								<td class="px-4 py-3">
									<Badge variant="default">{order.market}</Badge>
								</td>
								<td class="px-4 py-3">
									<Badge variant={order.side === 'BUY' ? 'success' : 'destructive'}>
										{getSideLabel(order.side)}
									</Badge>
								</td>
								<td class="px-4 py-3 text-right font-mono text-sm text-foreground">
									{order.price?.toFixed(2) || '-'}
								</td>
								<td class="px-4 py-3 text-right font-mono text-sm text-foreground">
									{order.qty || '-'}
								</td>
								<td class="px-4 py-3 text-right font-mono text-sm text-foreground">
									{order.fill_qty || '0'}
								</td>
								<td class="px-4 py-3 text-right font-mono text-sm text-foreground">
									{order.fill_price?.toFixed(2) || '-'}
								</td>
								<td class="px-4 py-3 text-center">
									<Badge variant={getStatusColor(order.status)}>
										{getStatusLabel(order.status)}
									</Badge>
								</td>
								<td class="px-4 py-3 text-sm text-muted-foreground">
									{formatTime(order.create_time)}
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</Card>
</div>
