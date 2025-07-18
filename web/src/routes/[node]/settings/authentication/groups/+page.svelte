<script lang="ts">
	import { addUsersToGroup, createGroup, deleteGroup, listGroups } from '$lib/api/auth/groups';
	import { listUsers } from '$lib/api/auth/local';
	import AlertDialog from '$lib/components/custom/Dialog/Alert.svelte';
	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import Search from '$lib/components/custom/TreeTable/Search.svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	import CustomComboBox from '$lib/components/ui/custom-input/combobox.svelte';
	import CustomValueInput from '$lib/components/ui/custom-input/value.svelte';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import type { Group, User } from '$lib/types/auth';
	import type { Column, Row } from '$lib/types/components/tree-table';
	import { handleAPIError, updateCache } from '$lib/utils/http';
	import { convertDbTime } from '$lib/utils/time';

	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import { toast } from 'svelte-sonner';
	import type { CellComponent } from 'tabulator-tables';

	interface Data {
		users: User[];
		groups: Group[];
	}

	let { data }: { data: Data } = $props();

	const results = useQueries([
		{
			queryKey: ['users'],
			queryFn: async () => {
				return (await listUsers()) as User[];
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.users,
			onSuccess: (data: User[]) => {
				updateCache('users', data);
			}
		},
		{
			queryKey: ['groups'],
			queryFn: async () => {
				return (await listGroups()) as Group[];
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.groups,
			onSuccess: (data: Group[]) => {
				updateCache('groups', data);
			}
		}
	]);

	let users = $derived($results[0].data as User[]);
	let groups = $derived($results[1].data as Group[]);
	let usersOptions = $derived.by(() => {
		return users.map((user) => ({
			label: user.username,
			value: user.username
		}));
	});

	let options = {
		create: {
			open: false,
			name: '',
			users: {
				open: false,
				value: [] as string[],
				data: (() => $state.snapshot(usersOptions))()
			}
		},
		delete: {
			open: false,
			id: 0
		},
		addUsers: {
			open: false,
			combobox: {
				open: false,
				value: [] as string[],
				data: (() => $state.snapshot(usersOptions))()
			}
		}
	};

	let properties = $state(options);

	async function onCreate() {
		let error = '';

		if (!properties.create.name.trim() || properties.create.users.value.length === 0) {
			error = 'Name and users are required';
		} else if (groups.some((g) => g.name === properties.create.name.trim())) {
			error = 'Group name already exists';
		}

		if (error) {
			toast.error(error, {
				position: 'bottom-center'
			});
			return;
		}

		const response = await createGroup(
			properties.create.name.trim(),
			properties.create.users.value
		);

		if (response.error) {
			handleAPIError(response);
			toast.error('Failed to create group', {
				position: 'bottom-center'
			});
			return;
		} else {
			toast.success('Group created', {
				position: 'bottom-center'
			});

			properties.create.open = false;
			properties.create.name = '';
			properties.create.users.value = [];
		}
	}

	async function onAddUsers() {
		if (properties.addUsers.combobox.value.length === 0) {
			toast.error('No users selected', {
				position: 'bottom-center'
			});
			return;
		}

		const response = await addUsersToGroup(
			properties.addUsers.combobox.value,
			activeRow ? activeRow.name : ''
		);

		if (response.status === 'error') {
			handleAPIError(response);
			toast.error('Failed to add users to group', {
				position: 'bottom-center'
			});
			return;
		} else {
			toast.success('Users added to group', {
				position: 'bottom-center'
			});

			properties.addUsers.open = false;
			properties.addUsers.combobox.value = [];
		}
	}

	function generateTableData(users: User[], groups: Group[]): { rows: Row[]; columns: Column[] } {
		const columns: Column[] = [
			{
				field: 'id',
				title: 'ID',
				visible: false
			},
			{
				field: 'name',
				title: 'Name'
			},
			{
				field: 'createdAt',
				title: 'Created At',
				formatter: (cell: CellComponent) => {
					const value = cell.getValue();
					return convertDbTime(value);
				}
			}
		];

		const rows: Row[] = [];

		for (const group of groups) {
			rows.push({
				id: group.id,
				name: group.name,
				createdAt: group.createdAt,
				user: false,
				children: group.users?.map((user) => ({
					id: user.id,
					name: user.username,
					createdAt: user.createdAt,
					user: true
				}))
			});
		}

		return {
			columns,
			rows
		};
	}

	let tableData = $derived(generateTableData(users, groups));
	let query: string = $state('');
	let activeRows: Row[] | null = $state(null);
	let activeRow: Row | null = $derived(activeRows ? (activeRows[0] as Row) : ({} as Row));
</script>

{#snippet button(type: string)}
	{#if activeRows && activeRows.length === 1 && !activeRows[0].user}
		{#if type === 'delete'}
			<Button
				onclick={() => {
					properties.delete.open = !properties.delete.open;
					properties.delete.id = activeRows ? (activeRows[0].id as number) : 0;
				}}
				size="sm"
				variant="outline"
				class="h-6.5"
			>
				<div class="flex items-center">
					<Icon icon="mdi:delete" class="mr-1 h-4 w-4" />
					<span>Delete</span>
				</div>
			</Button>
		{/if}

		{#if type === 'add-users'}
			<Button
				onclick={() => {
					properties.addUsers.open = !properties.addUsers.open;
					if (activeRows) {
						properties.addUsers.combobox.value =
							activeRows[0].children?.map((user) => user.name) || [];
					}
				}}
				size="sm"
				variant="outline"
				class="h-6.5"
			>
				<div class="flex items-center">
					<Icon icon="material-symbols:group-add" class="mr-1 h-4 w-4" />
					<span>Add Users</span>
				</div>
			</Button>
		{/if}
	{/if}
{/snippet}

<div class="flex h-full flex-col overflow-hidden">
	<div class="flex h-10 w-full items-center gap-2 border-b p-2">
		<Search bind:query />
		<Button
			onclick={() => (properties.create.open = !properties.create.open)}
			size="sm"
			class="h-6"
		>
			<div class="flex items-center">
				<Icon icon="gg:add" class="mr-1 h-4 w-4" />
				<span>New</span>
			</div>
		</Button>

		{@render button('add-users')}
		{@render button('delete')}
	</div>

	<TreeTable
		data={tableData}
		name={'tt-groups'}
		bind:parentActiveRow={activeRows}
		multipleSelect={false}
		bind:query
	/>
</div>

{#if properties.create.open}
	<Dialog.Root bind:open={properties.create.open}>
		<Dialog.Content
			class="sm:max-w-[425px]"
			onInteractOutside={(e) => e.preventDefault()}
			onEscapeKeydown={(e) => e.preventDefault()}
		>
			<Dialog.Header>
				<Dialog.Title class="flex items-center justify-between">
					<div class="flex items-center gap-2">
						<Icon icon="mdi:account-group" class="h-5 w-5" />
						<span>New Group</span>
					</div>
					<div class="flex items-center gap-0.5">
						<Button
							size="sm"
							variant="link"
							title={'Reset'}
							class="h-4"
							onclick={() => {
								properties = options;
							}}
						>
							<Icon icon="radix-icons:reset" class="pointer-events-none h-4 w-4" />
							<span class="sr-only">{'Reset'}</span>
						</Button>
						<Button
							size="sm"
							variant="link"
							class="h-4"
							title={'Close'}
							onclick={() => {
								properties = options;
								properties.create.open = false;
							}}
						>
							<Icon icon="material-symbols:close-rounded" class="pointer-events-none h-4 w-4" />
							<span class="sr-only">{'Close'}</span>
						</Button>
					</div>
				</Dialog.Title>
			</Dialog.Header>

			<CustomValueInput
				label={'Name'}
				placeholder="c-level"
				bind:value={properties.create.name}
				classes="flex-1 space-y-1.5"
			/>

			<CustomComboBox
				bind:open={properties.create.users.open}
				bind:value={properties.create.users.value}
				data={properties.create.users.data}
				onValueChange={(v) => {
					properties.create.users.value = v as string[];
				}}
				placeholder={'Select users'}
				multiple={true}
				width="w-full"
			/>

			<Dialog.Footer class="flex justify-end">
				<div class="flex w-full items-center justify-end gap-2">
					<Button onclick={() => onCreate()} type="submit" size="sm">{'Create'}</Button>
				</div>
			</Dialog.Footer>
		</Dialog.Content>
	</Dialog.Root>
{/if}

{#if properties.addUsers.open}
	<Dialog.Root bind:open={properties.addUsers.open}>
		<Dialog.Content
			class="sm:max-w-[425px]"
			onInteractOutside={(e) => e.preventDefault()}
			onEscapeKeydown={(e) => e.preventDefault()}
		>
			<Dialog.Header>
				<Dialog.Title class="flex items-center justify-between">
					<div class="flex items-center gap-2">
						<Icon icon="material-symbols:group-add" class="h-5 w-5" />
						<span>Add Users</span>
					</div>
					<div class="flex items-center gap-0.5">
						<Button
							size="sm"
							variant="link"
							title={'Reset'}
							class="h-4"
							onclick={() => {
								properties = options;
							}}
						>
							<Icon icon="radix-icons:reset" class="pointer-events-none h-4 w-4" />
							<span class="sr-only">{'Reset'}</span>
						</Button>
						<Button
							size="sm"
							variant="link"
							class="h-4"
							title={'Close'}
							onclick={() => {
								properties = options;
								properties.addUsers.open = false;
							}}
						>
							<Icon icon="material-symbols:close-rounded" class="pointer-events-none h-4 w-4" />
							<span class="sr-only">{'Close'}</span>
						</Button>
					</div>
				</Dialog.Title>
			</Dialog.Header>

			<CustomComboBox
				bind:open={properties.addUsers.combobox.open}
				bind:value={properties.addUsers.combobox.value}
				data={properties.addUsers.combobox.data}
				onValueChange={(v) => {
					properties.addUsers.combobox.value = v as string[];
				}}
				placeholder={'Select users'}
				multiple={true}
				width="w-full"
			/>

			<Dialog.Footer class="flex justify-end">
				<div class="flex w-full items-center justify-end gap-2">
					<Button onclick={() => onAddUsers()} type="submit" size="sm">{'Add Users'}</Button>
				</div>
			</Dialog.Footer>
		</Dialog.Content>
	</Dialog.Root>
{/if}

<AlertDialog
	open={properties.delete.open}
	names={{ parent: 'group', element: activeRow?.name || '' }}
	actions={{
		onConfirm: async () => {
			const result = await deleteGroup(properties.delete.id);
			if (result.status === 'error') {
				handleAPIError(result);
				toast.error('Failed to delete group', {
					position: 'bottom-center'
				});
				return;
			} else {
				toast.success('Group deleted', {
					position: 'bottom-center'
				});

				properties.delete.open = false;
				properties.delete.id = 0;
			}
		},
		onCancel: () => {
			properties.delete.open = false;
			properties.delete.id = 0;
		}
	}}
></AlertDialog>
