<script setup lang="ts">
import { type Album } from "@/api";
import { RouterLink } from "vue-router";

defineProps<{ modelValue: Album[] }>();
</script>

<template>
    <RouterLink v-for="album in modelValue" :key="album.Id" :to="'/albums/' + album.Id">
        <div>{{ album.Name }}</div>
        <div>by {{ album.Artist }}</div>
    </RouterLink>
</template>
