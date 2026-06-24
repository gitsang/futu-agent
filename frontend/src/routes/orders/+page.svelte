<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import type { Order } from '$lib/types';
	import { cn } from '$lib/utils';
	import { Card, Badge, LoadingSpinner, Button } from '$lib/components';
	import { selectedMarket } from '$lib/stores';

	let orders = $state<Order[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	onMount(async () => {
		try {
			orders = await api.getOrders();
		} catch (e) {
			error = e instanceof Error ? e.message : '加载失败';
		} finally {
			loading = false;
		}
	});

	let filteredOrders = $derived(
		$selectedMarket === 'ALL'
			? orders
			: orders.filter(o => o.market === $selectedMarket)
	);

	function getStatusColor(status: string): 'success' | 'destructive' | 'warning' | 'default' {
		switch (status) {
			case 'FILLED': return 'success';
			case 'CANCELLED': return 'destructive';
			case 'FAILED': return 'destructive';
			case 'PENDING': return 'warning';
			case 'SUBMITTED': return 'warning';
			default: return 'default';
		}
	}

	function getStatusLabel(status: string): string {
		switch (status) {
			case 'PENDING': return '待提交';
			case 'SUBMITTED': return '已提交';
			case 'FILLED': return '已成交';
			case 'CANCELLED': return '已撤单';
			case 'FAILED': return '失败';
			case 'EXPIRED': return '已过期';
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
			<h1 class="text-2xl font-semibold text-text-primary">订单管理</h1>
			<p class="text-sm text-text-secondary">查看所有交易订单状态</p>
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

	<Card padding={false}>
		<div class="overflow-x-auto">
			<table class="w-full">
				<thead>
					<tr class="border-b border-border">
						<th class="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">订单号</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">股票</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">市场</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">方向</th>
						<th class="px-4 py-3 text-right text-xs font-medium text-text-muted uppercase tracking-wider">委托价</th>
						<th class="px-4 py-3 text-right text-xs font-medium text-text-muted uppercase tracking-wider">委托量</th>
						<th class="px-4 py-3 text-right text-xs font-medium text-text-muted uppercase tracking-wider">成交量</th>
						<th class="px-4 py-3 text-right text-xs font-medium text-text-muted uppercase tracking-wider">成交价</th>
						<th class="px-4 py-3 text-center text-xs font-medium text-text-muted uppercase tracking-wider">状态</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">下单时间</th>
					</tr>
				</thead>
				<tbody>
					{#if loading}
						{#each Array(3) as _}
							<tr class="border-b border-border-subtle">
								{#each Array(10) as _}
									<td class="px-4 py-3">
										<div class="h-4 w-16 animate-pulse rounded bg-surface-elevated"></div>
									</td>
								{/each}
							</tr>
						{/each}
					{:else if filteredOrders.length === 0}
						<tr>
							<td colspan="10" class="px-4 py-12 text-center text-text-muted">
								暂无订单数据
							</td>
						</tr>
					{:else}
						{#each filteredOrders as order}
							<tr class="border-b border-border-subtle transition-colors hover:bg-surface-hover">
								<td class="px-4 py-3">
									<div class="font-mono text-xs text-text-muted">{order.order_id?.slice(-8) || '-'}</div>
								</td>
								<td class="px-4 py-3">
									<div class="font-medium text-text-primary">{order.code}</div>
									<div class="text-xs text-text-muted">{order.name}</div>
								</td>
								<td class="px-4 py-3">
									<Badge variant="default">{order.market}</Badge>
								</td>
								<td class="px-4 py-3">
									<Badge variant={order.side === 'BUY' ? 'success' : 'destructive'}>
										{getSideLabel(order.side)}
									</Badge>
								</td>
								<td class="px-4 py-3 text-right font-mono text-sm text-text-primary">
									{order.price?.toFixed(2) || '-'}
								</td>
								<td class="px-4 py-3 text-right font-mono text-sm text-text-primary">
									{order.qty || '-'}
								</td>
								<td class="px-4 py-3 text-right font-mono text-sm text-text-primary">
									{order.fill_qty || '0'}
								</td>
								<td class="px-4 py-3 text-right font-mono text-sm text-text-primary">
									{order.fill_price?.toFixed(2) || '-'}
								</td>
								<td class="px-4 py-3 text-center">
									<Badge variant={getStatusColor(order.status)}>
										{getStatusLabel(order.status)}
									</Badge>
								</td>
								<td class="px-4 py-3 text-sm text-text-muted">
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
