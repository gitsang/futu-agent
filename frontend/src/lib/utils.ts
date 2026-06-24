export function formatCurrency(value: number, currency = 'CNY'): string {
	return new Intl.NumberFormat('zh-CN', {
		style: 'currency',
		currency,
		minimumFractionDigits: 2,
		maximumFractionDigits: 2
	}).format(value);
}

export function formatNumber(value: number, decimals = 2): string {
	return new Intl.NumberFormat('zh-CN', {
		minimumFractionDigits: decimals,
		maximumFractionDigits: decimals
	}).format(value);
}

export function formatPercent(value: number): string {
	return `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`;
}

export function formatDate(dateStr: string): string {
	const date = new Date(dateStr);
	return new Intl.DateTimeFormat('zh-CN', {
		year: 'numeric',
		month: '2-digit',
		day: '2-digit',
		hour: '2-digit',
		minute: '2-digit'
	}).format(date);
}

export function cn(...classes: (string | boolean | undefined | null)[]): string {
	return classes.filter(Boolean).join(' ');
}
