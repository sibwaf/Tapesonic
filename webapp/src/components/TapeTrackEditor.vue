<script setup lang="ts">
import { type TapeTrack } from "@/api";

defineProps<{ modelValue: TapeTrack }>();
</script>

<template>
    <tr>
        <td>
            <input type="text" v-model="modelValue.Artist">
        </td>
        <td>
            <input type="text" v-model="modelValue.Title">
        </td>
    </tr>
</template>
