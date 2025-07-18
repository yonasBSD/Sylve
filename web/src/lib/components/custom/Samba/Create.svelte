<script lang="ts">
	import { createSambaShare } from '$lib/api/samba/share';
	import Button from '$lib/components/ui/button/button.svelte';
	import CustomComboBox from '$lib/components/ui/custom-input/combobox.svelte';
	import CustomValueInput from '$lib/components/ui/custom-input/value.svelte';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import type { Group } from '$lib/types/auth';
	import type { SambaShare } from '$lib/types/samba/shares';
	import type { Dataset } from '$lib/types/zfs/dataset';
	import Icon from '@iconify/svelte';
	import { toast } from 'svelte-sonner';

	interface Props {
		open: boolean;
		shares: SambaShare[];
		datasets: Dataset[];
		groups: Group[];
	}

	let { open = $bindable(), shares, datasets, groups }: Props = $props();

	let options = {
		name: '',
		dataset: {
			combobox: {
				open: false,
				value: '',
				options: datasets
					.filter(
						(dataset) =>
							dataset.properties.mountpoint !== '-' &&
							dataset.properties.mountpoint !== null &&
							dataset.properties.mountpoint !== '' &&
							dataset.properties.mountpoint !== '/' &&
							dataset.properties.mounted
					)
					.map((dataset) => ({
						label: dataset.name,
						value: dataset.properties.guid ? dataset.properties.guid : dataset.name
					}))
			}
		},
		readOnlyGroups: {
			combobox: {
				open: false,
				value: [] as string[],
				options: groups.map((group) => ({
					label: group.name,
					value: group.name
				}))
			}
		},
		writeableGroups: {
			combobox: {
				open: false,
				value: [] as string[],
				options: groups.map((group) => ({
					label: group.name,
					value: group.name
				}))
			}
		},
		createMask: '0664',
		directoryMask: '2775'
	};

	let properties = $state(options);

	async function create() {
		let error = '';

		if (shares.some((share) => share.name === properties.name)) {
			error = 'Share name already exists';
		}

		if (properties.name === '') {
			error = 'Name is required';
		} else if (properties.dataset.combobox.value === '') {
			error = 'Dataset is required';
		} else if (
			properties.readOnlyGroups.combobox.value.length === 0 &&
			properties.writeableGroups.combobox.value.length === 0
		) {
			error = 'No groups selected';
		}

		if (
			properties.readOnlyGroups.combobox.value.some((group) =>
				properties.writeableGroups.combobox.value.includes(group)
			)
		) {
			error = 'Share cannot have overlapping groups';
		}

		if (error) {
			toast.error(error, {
				position: 'bottom-center'
			});
			return;
		}

		const response = await createSambaShare(
			properties.name,
			properties.dataset.combobox.value,
			properties.readOnlyGroups.combobox.value,
			properties.writeableGroups.combobox.value,
			properties.createMask,
			properties.directoryMask
		);

		if (response.status === 'error') {
			toast.error('Failed to create Samba share', {
				position: 'bottom-center'
			});
			return;
		}

		toast.success('Samba share created', {
			position: 'bottom-center'
		});

		open = false;
		properties = options;
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content
		class="flex flex-col p-5"
		onInteractOutside={() => {
			properties = options;
			open = false;
		}}
	>
		<Dialog.Header class="p-0">
			<Dialog.Title class="flex justify-between gap-1 text-left">
				<div class="flex items-center gap-2">
					<Icon icon="mdi:folder-network" class="h-6 w-6" />
					<span>Create Samba Share</span>
				</div>

				<div class="flex items-center gap-0.5">
					<Button
						size="sm"
						variant="link"
						class="h-4"
						title={'Reset'}
						onclick={() => {
							properties = options;
						}}
					>
						<Icon icon="radix-icons:reset" class="pointer-events-none h-4 w-4" />
						<span class="sr-only">Reset</span>
					</Button>
					<Button
						size="sm"
						variant="link"
						class="h-4"
						title={'Close'}
						onclick={() => {
							open = false;
							properties = options;
						}}
					>
						<Icon icon="material-symbols:close-rounded" class="pointer-events-none h-4 w-4" />
						<span class="sr-only">Close</span>
					</Button>
				</div>
			</Dialog.Title>
		</Dialog.Header>

		<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
			<CustomValueInput
				label={'Name'}
				placeholder="share"
				bind:value={properties.name}
				classes="flex-1 space-y-1.5"
			/>

			<CustomComboBox
				label={'Dataset'}
				placeholder="Select dataset"
				bind:open={properties.dataset.combobox.open}
				bind:value={properties.dataset.combobox.value}
				data={properties.dataset.combobox.options}
				multiple={false}
				width="w-full"
			/>
		</div>

		<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
			<CustomComboBox
				label={'Read-Only Groups'}
				placeholder="Select groups"
				bind:open={properties.readOnlyGroups.combobox.open}
				bind:value={properties.readOnlyGroups.combobox.value}
				data={properties.readOnlyGroups.combobox.options}
				multiple={true}
				width="w-full"
			/>

			<CustomComboBox
				label={'Writeable Groups'}
				placeholder="Select groups"
				bind:open={properties.writeableGroups.combobox.open}
				bind:value={properties.writeableGroups.combobox.value}
				data={properties.writeableGroups.combobox.options}
				multiple={true}
				width="w-full"
			/>
		</div>

		<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
			<CustomValueInput
				label={'Create Mask'}
				placeholder="0664"
				bind:value={properties.createMask}
				classes="flex-1 space-y-1.5"
			/>

			<CustomValueInput
				label={'Directory Mask'}
				placeholder="2775"
				bind:value={properties.directoryMask}
				classes="flex-1 space-y-1.5"
			/>
		</div>

		<Dialog.Footer class="mt-4">
			<div class="flex items-center justify-end space-x-4">
				<Button
					size="sm"
					type="button"
					class="h-8 w-full lg:w-28"
					onclick={() => {
						create();
					}}
				>
					Create
				</Button>
			</div>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
