<script>
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	onMount(() => goto('/campaigns', { replaceState: true }));
</script>
