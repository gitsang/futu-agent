<script lang="ts">
	import { page } from '$app/stores';
	import { cn } from '$lib/utils';
	import StatusIndicator from '$lib/components/StatusIndicator.svelte';
	import type { Snippet } from 'svelte';

	let { children }: { children: Snippet } = $props();

	const navItems = [
		{ href: '/', label: '仪表盘', icon: '📊' },
		{ href: '/decisions', label: '决策日志', icon: '🤖' },
		{ href: '/positions', label: '持仓管理', icon: '📈' },
		{ href: '/config', label: '系统配置', icon: '⚙️' }
	];

	let sidebarOpen = $state(false);
</script>

<div class="flex h-screen overflow-hidden">
	{#if sidebarOpen}
		<div
			class="fixed inset-0 z-40 bg-black/50 lg:hidden"
			onclick={() => sidebarOpen = false}
			role="presentation"
		></div>
	{/if}

	<aside
		class={cn(
			'fixed inset-y-0 left-0 z-50 w-64 bg-surface border-r border-border flex flex-col transition-transform duration-300 lg:static lg:translate-x-0',
			sidebarOpen ? 'translate-x-0' : '-translate-x-full'
		)}
	>
		<div class="flex items-center gap-3 px-6 py-5 border-b border-border">
			<div class="flex h-10 w-10 items-center justify-center rounded-xl bg-accent/10 text-accent">
				<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
				</svg>
			</div>
			<div>
				<h1 class="text-lg font-semibold text-text-primary">Futu Agent</h1>
				<p class="text-xs text-text-muted">智能交易系统</p>
			</div>
		</div>

		<nav class="flex-1 px-3 py-4 space-y-1">
			{#each navItems as item}
				{@const isActive = $page.url.pathname === item.href}
				<a
					href={item.href}
					class={cn(
						'flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-all duration-200',
						isActive
							? 'bg-accent/10 text-accent'
							: 'text-text-secondary hover:text-text-primary hover:bg-surface-hover'
					)}
					onclick={() => sidebarOpen = false}
				>
					<span class="text-lg">{item.icon}</span>
					<span>{item.label}</span>
				</a>
			{/each}
		</nav>

		<div class="border-t border-border px-4 py-4 space-y-2">
			<StatusIndicator status="online" label="服务运行中" />
			<StatusIndicator status="online" label="数据库已连接" />
			<StatusIndicator status="online" label="Futu OpenD 已连接" />
		</div>
	</aside>

	<div class="flex flex-1 flex-col overflow-hidden">
		<header class="flex items-center justify-between border-b border-border bg-surface px-4 py-3 lg:px-6">
			<button
				class="rounded-lg p-2 text-text-secondary hover:text-text-primary hover:bg-surface-hover lg:hidden"
				onclick={() => sidebarOpen = !sidebarOpen}
				aria-label="切换侧边栏"
			>
				<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
				</svg>
			</button>

			<div class="flex items-center gap-4">
				<div class="hidden sm:flex items-center gap-2 rounded-lg bg-surface-elevated px-3 py-2">
					<svg class="h-4 w-4 text-text-muted" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
					</svg>
					<input
						type="text"
						placeholder="搜索..."
						class="bg-transparent text-sm text-text-primary placeholder-text-muted outline-none w-48"
					/>
				</div>

				<button class="relative rounded-lg p-2 text-text-secondary hover:text-text-primary hover:bg-surface-hover" aria-label="通知">
					<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
					</svg>
					<span class="absolute right-1.5 top-1.5 h-2 w-2 rounded-full bg-accent"></span>
				</button>
			</div>
		</header>

		<main class="flex-1 overflow-y-auto p-4 lg:p-6">
			{@render children()}
		</main>
	</div>
</div>
