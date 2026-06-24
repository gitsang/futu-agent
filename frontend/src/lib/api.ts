import { browser } from '$app/environment';
import type { AccountFunds, Agent, Decision, Position, SystemConfig, SystemStatus } from './types';

const BASE_URL = browser ? '' : 'http://localhost:8080';

async function fetchApi<T>(endpoint: string, options?: RequestInit): Promise<T> {
	const url = `${BASE_URL}/api${endpoint}`;
	const response = await fetch(url, {
		...options,
		headers: {
			'Content-Type': 'application/json',
			...options?.headers
		}
	});

	if (!response.ok) {
		throw new Error(`API error: ${response.status} ${response.statusText}`);
	}

	return response.json();
}

export const api = {
	getFunds: (market?: string) => fetchApi<AccountFunds>(`/account/funds${market ? `?market=${market}` : ''}`),
	getAllFunds: () => fetchApi<AccountFunds[]>('/account/funds/all'),
	getPositions: (market?: string) => fetchApi<Position[]>(`/account/positions${market ? `?market=${market}` : ''}`),

	getDecisions: (market?: string) => fetchApi<Decision[]>(`/decisions${market ? `?market=${market}` : ''}`),
	getDecision: (id: string) => fetchApi<Decision>(`/decisions/${id}`),

	getAgents: () => fetchApi<Agent[]>('/agents'),
	createAgent: (agent: Omit<Agent, 'id'>) =>
		fetchApi<Agent>('/agents', { method: 'POST', body: JSON.stringify(agent) }),
	updateAgent: (id: string, agent: Partial<Agent>) =>
		fetchApi<Agent>(`/agents/${id}`, { method: 'PUT', body: JSON.stringify(agent) }),
	deleteAgent: (id: string) =>
		fetchApi<void>(`/agents/${id}`, { method: 'DELETE' }),

	getConfig: () => fetchApi<SystemConfig>('/config'),
	getStatus: () => fetchApi<SystemStatus>('/status')
};
