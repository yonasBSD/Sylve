<script lang="ts">
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { generateNanoId } from '$lib/utils/string';

	interface Props {
		label: string;
		checked: boolean;
		classes: string;
	}

	let {
		label = '',
		checked = $bindable(false),
		classes = 'flex items-center gap-2'
	}: Props = $props();

	let nanoId = $state(generateNanoId(label));
</script>

<div class={classes}>
	<Checkbox id={nanoId} bind:checked aria-labelledby={label} />
	<Label
		id={nanoId}
		for={nanoId}
		class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
	>
		{label}
	</Label>
</div>
