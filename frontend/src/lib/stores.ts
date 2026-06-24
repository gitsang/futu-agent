import { writable, derived } from 'svelte/store';

export const MARKETS = [
	{ id: 'ALL', label: '全部', icon: '🌍', currency: 'CNY' },
	{ id: 'CN', label: 'A股', icon: '🇨🇳', currency: 'CNY' },
	{ id: 'HK', label: '港股', icon: '🇭🇰', currency: 'HKD' },
	{ id: 'US', label: '美股', icon: '🇺🇸', currency: 'USD' }
] as const;

export type MarketId = typeof MARKETS[number]['id'];

export const selectedMarket = writable<MarketId>('ALL');

export const currentMarket = derived(selectedMarket, ($selectedMarket) => {
	return MARKETS.find(m => m.id === $selectedMarket) || MARKETS[0];
});
