export interface AccountFunds {
	market: string;
	currency: string;
	total_assets: number;
	cash: number;
	market_value: number;
}

export interface Position {
	code: string;
	market: string;
	name: string;
	quantity: number;
	avg_cost: number;
	current_price: number;
	unrealized_pnl: number;
}

export interface Decision {
	id: string;
	agent_id: string;
	stock_code: string;
	market: string;
	action: string;
	quantity: number;
	price: number;
	reason: string;
	executed: boolean;
	created_at: string;
}

export interface Agent {
	id: string;
	agent_id: string;
	market: string;
	name: string;
	description: string;
	llm_model: string;
	enabled: boolean;
}

export interface SystemConfig {
	[key: string]: unknown;
}

export interface SystemStatus {
	server_status: string;
	database_status: string;
	futu_opend_status: string;
	trading_enabled: boolean;
	active_agents: number;
}
