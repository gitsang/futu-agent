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
	getFunds: () => fetchApi<AccountFunds>('/account/funds'),
	getPositions: () => fetchApi<Position[]>('/account/positions'),

	getDecisions: () => fetchApi<Decision[]>('/decisions'),
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
