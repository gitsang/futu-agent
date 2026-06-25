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

export function exportToCSV(data: Record<string, unknown>[], filename: string) {
	if (data.length === 0) return;

	const headers = Object.keys(data[0]);
	const csvContent = [
		headers.join(','),
		...data.map(row =>
			headers.map(header => {
				const value = row[header];
				if (value === null || value === undefined) return '';
				const stringValue = String(value);
				if (stringValue.includes(',') || stringValue.includes('"') || stringValue.includes('\n')) {
					return `"${stringValue.replace(/"/g, '""')}"`;
				}
				return stringValue;
			}).join(',')
		)
	].join('\n');

	const blob = new Blob(['\ufeff' + csvContent], { type: 'text/csv;charset=utf-8;' });
	const link = document.createElement('a');
	link.href = URL.createObjectURL(blob);
	link.download = `${filename}_${new Date().toISOString().slice(0, 10)}.csv`;
	link.click();
	URL.revokeObjectURL(link.href);
}
