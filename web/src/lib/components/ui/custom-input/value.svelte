<script lang="ts">
	import Input from '$lib/components/ui/input/input.svelte';
	import Label from '$lib/components/ui/label/label.svelte';
	import Textarea from '$lib/components/ui/textarea/textarea.svelte';
	import { generateNanoId } from '$lib/utils/string';
	import type { FullAutoFill } from 'svelte/elements';

	interface Props {
		label: string;
		value: string | number;
		placeholder: string;
		autocomplete?: FullAutoFill | null | undefined;
		classes: string;
		type?: string;
		textAreaCLasses?: string;
		disabled?: boolean;
	}

	let {
		value = $bindable(''),
		label = '',
		placeholder = '',
		autocomplete = 'off',
		classes = 'space-y-1',
		type = 'text',
		textAreaCLasses = 'min-h-56',
		disabled = false
	}: Props = $props();

	let nanoId = $state(generateNanoId(label));
</script>

<div class={`${classes}`}>
	{#if label}
		<Label for={nanoId}>{label}</Label>
	{/if}
	{#if type === 'textarea'}
		<Textarea
			class={textAreaCLasses}
			id={nanoId}
			{placeholder}
			{autocomplete}
			bind:value
			{disabled}
		/>
	{:else}
		<Input {type} id={nanoId} {placeholder} {autocomplete} bind:value {disabled} />
	{/if}
</div>
