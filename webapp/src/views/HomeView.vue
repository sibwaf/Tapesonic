<script setup lang="ts">
import AlbumGrid from "@/components/AlbumGrid.vue";
import TapeGrid from "@/components/TapeGrid.vue";
import TapeImporter from "@/components/TapeImporter.vue";
import PlaylistGrid from "@/components/PlaylistGrid.vue";
import { ref } from "vue";
import api, { type Album, type Playlist, type Tape } from "@/api";

enum State {
    LOADING,
    OK,
    ERROR,
}

const state = ref(State.LOADING);
const tapes = ref<Tape[]>([]);
const playlists = ref<Playlist[]>([]);
const albums = ref<Album[]>([]);

(async () => {
    try {
        state.value = State.LOADING;

        const tapesAsync = api.getAllTapes();
        const playlistsAsync = api.getAllPlaylists();
        const albumsAsync = api.getAllAlbums();

        tapes.value = await tapesAsync;
        playlists.value = await playlistsAsync;
        albums.value = await albumsAsync;

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

        <hr>

        <h1>Albums</h1>
        <AlbumGrid v-model="albums" />
    </div>
</template>
