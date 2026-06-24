import type { Handle } from '@sveltejs/kit';

const BACKEND_URL = 'http://127.0.0.1:9080';

export const handle: Handle = async ({ event, resolve }) => {
	if (event.url.pathname.startsWith('/api')) {
		const url = `${BACKEND_URL}${event.url.pathname}${event.url.search}`;
		
		const response = await fetch(url, {
			method: event.request.method,
			headers: event.request.headers,
			body: event.request.method !== 'GET' && event.request.method !== 'HEAD' 
				? await event.request.text() 
				: undefined
		});

		const headers = new Headers();
		response.headers.forEach((value, key) => {
			if (key !== 'transfer-encoding') {
				headers.set(key, value);
			}
		});

		return new Response(await response.text(), {
			status: response.status,
			statusText: response.statusText,
			headers
		});
	}

	return resolve(event);
};
