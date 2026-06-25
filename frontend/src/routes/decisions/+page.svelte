<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import type { Decision, PaginatedResponse } from '$lib/types';
	import { formatDate, cn, exportToCSV } from '$lib/utils';
	import { Card, Badge, LoadingSpinner, Button } from '$lib/components';
	import { selectedMarket } from '$lib/stores';

	let decisions = $state<Decision[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let selectedDecision = $state<Decision | null>(null);
	let filter = $state<'all' | 'buy' | 'sell' | 'hold'>('all');
	let currentPage = $state(1);
	let totalPages = $state(1);
	let totalItems = $state(0);
	const pageSize = 20;

	async function loadDecisions(page: number = 1) {
		loading = true;
		try {
			const result = await api.getDecisions($selectedMarket === 'ALL' ? undefined : $selectedMarket, page, pageSize);
			decisions = result.data || [];
			currentPage = result.page;
			totalPages = result.total_pages;
			totalItems = result.total;
		} catch (e) {
			error = e instanceof Error ? e.message : '加载失败';
		} finally {
			loading = false;
		}
	}

	onMount(() => loadDecisions());

	$effect(() => {
		if ($selectedMarket) {
			loadDecisions(1);
		}
	});

	function handleExport() {
		const exportData = filteredDecisions.map(d => ({
			'ID': d.id,
			'代理': d.agent_id,
			'股票代码': d.stock_code,
			'市场': d.market,
			'操作': getActionLabel(d.action),
			'数量': d.quantity,
			'价格': d.price,
			'原因': d.reason,
			'状态': d.executed ? '已执行' : '待执行',
			'时间': formatDate(d.created_at)
		}));
		exportToCSV(exportData, '决策日志');
	}

	let filteredDecisions = $derived(
		filter === 'all'
			? decisions
			: decisions.filter(d => d.action?.toLowerCase() === filter)
	);

	function goToPage(page: number) {
		if (page >= 1 && page <= totalPages) {
			loadDecisions(page);
		}
	}

	function getActionVariant(action: string) {
		switch (action?.toLowerCase()) {
			case 'buy': return 'success' as const;
			case 'sell': return 'destructive' as const;
			default: return 'default' as const;
		}
	}

	function getActionLabel(action: string) {
		switch (action?.toLowerCase()) {
			case 'buy': return '买入';
			case 'sell': return '卖出';
			default: return '持有';
		}
	}
</script>

<div class="space-y-6 animate-fade-in">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold text-foreground">决策日志</h1>
			<p class="text-sm text-muted-foreground">AI 交易决策记录与分析</p>
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

	<div class="flex gap-2">
		{#each [{ value: 'all', label: '全部' }, { value: 'buy', label: '买入' }, { value: 'sell', label: '卖出' }, { value: 'hold', label: '持有' }] as f}
			<button
				class={cn(
					'rounded-lg px-4 py-2 text-sm font-medium transition-all duration-200',
					filter === f.value
						? 'bg-primary text-primary-foreground'
						: 'bg-muted text-muted-foreground hover:text-foreground hover:bg-muted/80'
				)}
				onclick={() => filter = f.value as typeof filter}
			>
				{f.label}
			</button>
		{/each}
	</div>

	<div class="grid grid-cols-1 gap-4 lg:grid-cols-3">
		<div class="lg:col-span-2 space-y-3">
			{#if loading}
				<div class="flex justify-center py-12">
					<LoadingSpinner size="lg" />
				</div>
			{:else if filteredDecisions.length === 0}
				<Card>
					<div class="py-12 text-center text-muted-foreground">暂无决策记录</div>
				</Card>
			{:else}
				{#each filteredDecisions as decision}
					<button
						class={cn(
							'w-full text-left rounded-xl border p-4 transition-all duration-200 cursor-pointer',
							selectedDecision?.id === decision.id
								? 'border-primary bg-primary/5'
								: 'border-border bg-card hover:border-border/80 hover:bg-muted'
						)}
						onclick={() => selectedDecision = decision}
					>
						<div class="flex items-center justify-between mb-2">
							<div class="flex items-center gap-3">
								<span class="font-semibold text-foreground">{decision.stock_code}</span>
								<span class="text-xs text-muted-foreground">{decision.market}</span>
								<Badge variant={getActionVariant(decision.action)}>
									{getActionLabel(decision.action)}
								</Badge>
							</div>
							<Badge variant={decision.executed ? 'success' : 'warning'}>
								{decision.executed ? '已执行' : '待执行'}
							</Badge>
						</div>
						<p class="text-sm text-muted-foreground line-clamp-2 mb-2">{decision.reason}</p>
						<div class="flex items-center justify-between text-xs text-muted-foreground">
							<span>数量: {decision.quantity} · 价格: {decision.price}</span>
							<span>{formatDate(decision.created_at)}</span>
						</div>
					</button>
				{/each}

				{#if totalPages > 1}
					<div class="flex items-center justify-between pt-4">
						<div class="text-sm text-muted-foreground">
							共 {totalItems} 条记录，第 {currentPage}/{totalPages} 页
						</div>
						<div class="flex gap-2">
							<Button
								variant="secondary"
								size="sm"
								disabled={currentPage <= 1}
								onclick={() => goToPage(currentPage - 1)}
							>
								上一页
							</Button>
							<Button
								variant="secondary"
								size="sm"
								disabled={currentPage >= totalPages}
								onclick={() => goToPage(currentPage + 1)}
							>
								下一页
							</Button>
						</div>
					</div>
				{/if}
			{/if}
		</div>

		<div class="lg:col-span-1">
			{#if selectedDecision}
				<Card class="sticky top-6">
					<h3 class="text-lg font-semibold text-foreground mb-4">决策详情</h3>

					<div class="space-y-4">
						<div>
							<span class="text-xs text-muted-foreground">股票代码</span>
							<div class="text-sm font-medium text-foreground">{selectedDecision.stock_code}</div>
						</div>
						<div>
							<span class="text-xs text-muted-foreground">市场</span>
							<div class="text-sm text-foreground">{selectedDecision.market}</div>
						</div>
						<div>
							<span class="text-xs text-muted-foreground">操作</span>
							<div>
								<Badge variant={getActionVariant(selectedDecision.action)}>
									{getActionLabel(selectedDecision.action)}
								</Badge>
							</div>
						</div>
						<div class="grid grid-cols-2 gap-4">
							<div>
								<span class="text-xs text-muted-foreground">数量</span>
								<div class="text-sm font-mono text-foreground">{selectedDecision.quantity}</div>
							</div>
							<div>
								<span class="text-xs text-muted-foreground">价格</span>
								<div class="text-sm font-mono text-foreground">{selectedDecision.price}</div>
							</div>
						</div>
						<div>
							<span class="text-xs text-muted-foreground">决策原因</span>
							<p class="mt-1 text-sm text-muted-foreground leading-relaxed">{selectedDecision.reason}</p>
						</div>
						<div>
							<span class="text-xs text-muted-foreground">执行状态</span>
							<div>
								<Badge variant={selectedDecision.executed ? 'success' : 'warning'}>
									{selectedDecision.executed ? '已执行' : '待执行'}
								</Badge>
							</div>
						</div>
						<div>
							<span class="text-xs text-muted-foreground">创建时间</span>
							<div class="text-sm text-foreground">{formatDate(selectedDecision.created_at)}</div>
						</div>
					</div>
				</Card>
			{:else}
				<Card>
					<div class="py-12 text-center text-muted-foreground">
						<svg class="mx-auto h-12 w-12 mb-3 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
						</svg>
						<p class="text-sm">选择一条决策查看详情</p>
					</div>
				</Card>
			{/if}
		</div>
	</div>
</div>
