<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import type { Agent, SystemConfig, SystemStatus } from '$lib/types';
	import { cn } from '$lib/utils';
	import { Card, Badge, LoadingSpinner, Button, StatusIndicator } from '$lib/components';
	import { selectedMarket } from '$lib/stores';

	let status = $state<SystemStatus | null>(null);
	let agents = $state<Agent[]>([]);
	let config = $state<SystemConfig | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	onMount(async () => {
		try {
			const [statusData, agentsData, configData] = await Promise.all([
				api.getStatus(),
				api.getAgents(),
				api.getConfig()
			]);
			status = statusData;
			agents = agentsData;
			config = configData;
		} catch (e) {
			error = e instanceof Error ? e.message : '加载失败';
		} finally {
			loading = false;
		}
	});

	let filteredAgents = $derived(
		$selectedMarket === 'ALL'
			? agents
			: agents.filter(a => a.market === $selectedMarket)
	);

	async function toggleAgent(agent: Agent) {
		const action = agent.enabled ? '禁用' : '启用';
		const confirmed = confirm(`确定要${action}代理 "${agent.name}" 吗？`);
		if (!confirmed) return;

		try {
			await api.updateAgent(agent.id, { enabled: !agent.enabled });
			agents = agents.map((a) =>
				a.id === agent.id ? { ...a, enabled: !a.enabled } : a
			);
		} catch (e) {
			error = e instanceof Error ? e.message : '更新失败';
		}
	}

	function getStatusColor(statusStr: string) {
		switch (statusStr.toLowerCase()) {
			case 'ok':
			case 'online':
			case 'connected':
				return 'online' as const;
			case 'error':
			case 'offline':
			case 'disconnected':
				return 'offline' as const;
			case 'warning':
			case 'degraded':
				return 'warning' as const;
			default:
				return 'unknown' as const;
		}
	}
</script>

<div class="space-y-6 animate-fade-in">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold text-text-primary">系统配置</h1>
			<p class="text-sm text-text-secondary">系统状态监控与代理管理</p>
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
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-3 text-loss">
					<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
					</svg>
					<span class="text-sm">{error}</span>
				</div>
				<button class="text-text-muted hover:text-text-primary" onclick={() => error = null}>✕</button>
			</div>
		</Card>
	{/if}

	<Card>
		<h2 class="text-lg font-semibold text-text-primary mb-4">系统状态</h2>
		{#if loading}
			<div class="flex justify-center py-8">
				<LoadingSpinner />
			</div>
		{:else if status}
			<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
				<div class="flex items-center justify-between rounded-lg bg-surface-elevated p-4">
					<div class="flex items-center gap-3">
						<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-accent/10">
							<svg class="h-5 w-5 text-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M12 5l7 7-7 7" />
							</svg>
						</div>
						<div>
							<div class="text-sm text-text-muted">服务器状态</div>
							<div class="font-medium text-text-primary">{status.server_status}</div>
						</div>
					</div>
					<StatusIndicator status={getStatusColor(status.server_status)} label="" />
				</div>

				<div class="flex items-center justify-between rounded-lg bg-surface-elevated p-4">
					<div class="flex items-center gap-3">
						<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-accent/10">
							<svg class="h-5 w-5 text-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
							</svg>
						</div>
						<div>
							<div class="text-sm text-text-muted">Futu OpenD</div>
							<div class="font-medium text-text-primary">{status.futu_opend_status}</div>
						</div>
					</div>
					<StatusIndicator status={getStatusColor(status.futu_opend_status)} label="" />
				</div>

				<div class="flex items-center justify-between rounded-lg bg-surface-elevated p-4">
					<div class="flex items-center gap-3">
						<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-accent/10">
							<svg class="h-5 w-5 text-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
							</svg>
						</div>
						<div>
							<div class="text-sm text-text-muted">交易状态</div>
							<div class="font-medium text-text-primary">
								{status.trading_enabled ? '已启用' : '已禁用'}
							</div>
						</div>
					</div>
					<Badge variant={status.trading_enabled ? 'success' : 'destructive'}>
						{status.trading_enabled ? '启用' : '禁用'}
					</Badge>
				</div>

				<div class="flex items-center justify-between rounded-lg bg-surface-elevated p-4">
					<div class="flex items-center gap-3">
						<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-accent/10">
							<svg class="h-5 w-5 text-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
							</svg>
						</div>
						<div>
							<div class="text-sm text-text-muted">活跃代理</div>
							<div class="font-medium text-text-primary">{status.active_agents} 个</div>
						</div>
					</div>
				</div>
			</div>
		{/if}
	</Card>

	<Card>
		<div class="flex items-center justify-between mb-4">
			<h2 class="text-lg font-semibold text-text-primary">交易代理</h2>
		</div>

		{#if loading}
			<div class="flex justify-center py-8">
				<LoadingSpinner />
			</div>
		{:else if filteredAgents.length === 0}
			<div class="py-8 text-center text-text-muted">暂无代理配置</div>
		{:else}
			<div class="space-y-3">
				{#each filteredAgents as agent}
					<div class="flex items-center justify-between rounded-lg bg-surface-elevated p-4 transition-colors hover:bg-surface-hover">
						<div class="flex items-center gap-4">
							<div class={cn(
								'flex h-10 w-10 items-center justify-center rounded-lg',
								agent.enabled ? 'bg-profit/10 text-profit' : 'bg-surface text-text-muted'
							)}>
								<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
								</svg>
							</div>
							<div>
								<div class="font-medium text-text-primary">{agent.name}</div>
								<div class="text-xs text-text-muted">{agent.description}</div>
								<div class="mt-1 flex items-center gap-2">
									<Badge variant="info">{agent.llm_model}</Badge>
									<span class="text-xs text-text-muted">ID: {agent.id}</span>
								</div>
							</div>
						</div>
						<div class="flex items-center gap-2">
							<button
								class={cn(
									'relative inline-flex h-6 w-11 items-center rounded-full transition-colors duration-200',
									agent.enabled ? 'bg-profit' : 'bg-border'
								)}
								onclick={() => toggleAgent(agent)}
								role="switch"
								aria-checked={agent.enabled}
								aria-label="{agent.enabled ? '禁用' : '启用'} {agent.name}"
							>
								<span
									class={cn(
										'inline-block h-4 w-4 rounded-full bg-white transition-transform duration-200',
										agent.enabled ? 'translate-x-6' : 'translate-x-1'
									)}
								></span>
							</button>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</Card>

	{#if config}
		<Card>
			<h2 class="text-lg font-semibold text-text-primary mb-4">系统配置</h2>
			<div class="rounded-lg bg-surface-elevated p-4 font-mono text-sm">
				<pre class="text-text-secondary overflow-x-auto">{JSON.stringify(config, null, 2)}</pre>
			</div>
		</Card>
	{/if}
</div>
