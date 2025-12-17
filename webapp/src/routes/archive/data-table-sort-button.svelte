<script lang="ts">
 import type { ComponentProps } from "svelte";
 import ArrowUpDownIcon from "@lucide/svelte/icons/arrow-up-down";
 import { Button } from "$lib/components/ui/button/index.js";

 type sortButtonProp={
  compProp:ComponentProps<typeof Button>,
  label:string
 }
 
 let {label,compProp}: sortButtonProp =
  $props();
let {variant = "ghost", ...restProps }=compProp
</script>
 
<Button {variant} {...restProps}>
 {label}
 <ArrowUpDownIcon class="ml-2" />
</Button>