import { APIResponseSchema, type APIResponse } from '$lib/types/common';
import { SwitchListSchema, type SwitchList } from '$lib/types/network/switch';
import { apiRequest } from '$lib/utils/http';

export async function getSwitches(): Promise<SwitchList> {
	return await apiRequest('/network/switch', SwitchListSchema, 'GET');
}

export async function createSwitch(
	name: string,
	mtu: number,
	vlan: number,
	address: string,
	address6: string,
	privateSw: boolean,
	dhcp: boolean,
	ports: string[]
): Promise<APIResponse> {
	const body = {
		name,
		mtu,
		vlan,
		address,
		address6,
		private: privateSw,
		ports,
		dhcp
	};

	return await apiRequest('/network/switch/standard', APIResponseSchema, 'POST', body);
}

export async function deleteSwitch(id: number): Promise<APIResponse> {
	return await apiRequest(`/network/switch/standard/${id}`, APIResponseSchema, 'DELETE');
}

export async function updateSwitch(
	id: number,
	mtu: number,
	vlan: number,
	address: string,
	address6: string,
	privateSw: boolean,
	ports: string[]
): Promise<APIResponse> {
	const body = {
		id,
		mtu,
		vlan,
		address,
		address6,
		private: privateSw,
		ports
	};

	return await apiRequest('/network/switch/standard', APIResponseSchema, 'PUT', body);
}
