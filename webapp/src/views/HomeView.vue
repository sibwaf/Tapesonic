<script setup lang="ts">
import TapeGrid from "@/components/TapeGrid.vue";
import TapeImporter from "@/components/TapeImporter.vue";
import PlaylistGrid from "@/components/PlaylistGrid.vue";
</script>

<template>
    <TapeImporter />
    <hr>
    <h1>Tapes</h1>
    <TapeGrid />
    <hr>
    <h1>Playlists</h1>
    <PlaylistGrid />
</template>
