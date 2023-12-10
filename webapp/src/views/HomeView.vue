<script setup lang="ts">
import TapeGrid from "@/components/TapeGrid.vue";
import TapeImporter from "@/components/TapeImporter.vue";
import PlaylistGrid from "@/components/PlaylistGrid.vue";
import { ref } from "vue";
import api, { type Playlist, type Tape } from "@/api";

enum State {
    LOADING,
    OK,
    ERROR,
}

const state = ref(State.LOADING);
const tapes = ref<Tape[]>([]);
const playlists = ref<Playlist[]>([]);

(async () => {
    try {
        state.value = State.LOADING;

        const tapesAsync = api.getAllTapes();
        const playlistsAsync = api.getAllPlaylists();

        tapes.value = await tapesAsync;
        playlists.value = await playlistsAsync;

        state.value = State.OK;
    } catch (e) {
        state.value = State.ERROR;
    }
})();
</script>

<template>
    <TapeImporter />

    <hr>

    <div v-if="state == State.LOADING">
        Loading...
    </div>
    <div v-else-if="state == State.ERROR">
        Failed to load data
    </div>
    <div v-else>
        <h1>Tapes</h1>
        <TapeGrid v-model="tapes" />

        <hr>

        <h1>Playlists</h1>
        <PlaylistGrid v-model="playlists" />
    </div>
</template>
