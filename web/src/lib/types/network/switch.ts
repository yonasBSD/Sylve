import { array, z } from 'zod';

export const NetworkPortSchema = z.object({
	id: z.number(),
	name: z.string(),
	switchId: z.number()
});

export const StandardSwitchSchema = z.object({
	id: z.number(),
	name: z.string(),
	mtu: z.number(),
	vlan: z.number(),
	private: z.boolean(),
	address: z.string(),
	address6: z.string(),
	ports: array(NetworkPortSchema).optional(),
	dhcp: z.boolean().optional(),
	slaac: z.boolean(),
	disableIPv6: z.boolean()
});

export const SwitchListSchema = z.object({
	standard: array(StandardSwitchSchema).optional()
});

export type StandardSwitch = z.infer<typeof StandardSwitchSchema>;
export type SwitchList = z.infer<typeof SwitchListSchema>;
