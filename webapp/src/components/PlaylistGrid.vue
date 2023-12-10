<script setup lang="ts">
import { type Playlist } from "@/api";
import { RouterLink } from "vue-router";

defineProps<{ modelValue: Playlist[] }>();
</script>

<template>
    <RouterLink v-for="playlist in modelValue" :key="playlist.Id" :to="'/playlists/' + playlist.Id">
        <div>{{ playlist.Name }}</div>
    </RouterLink>
</template>
