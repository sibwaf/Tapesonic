<script setup lang="ts">
import { type Tape } from "@/api";
import { RouterLink } from "vue-router";

defineProps<{ modelValue: Tape[] }>();
</script>

<template>
    <RouterLink v-for="tape in modelValue" :key="tape.Id" :to="'/tapes/' + tape.Id">
        <div>{{ tape.Name }}</div>
        <div>by {{ tape.AuthorName }}</div>
    </RouterLink>
</template>
