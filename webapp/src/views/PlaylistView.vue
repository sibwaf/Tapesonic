<script setup lang="ts">
import api, { type Playlist } from "@/api";
import { useRoute } from "vue-router";
import { ref } from "vue";

enum State {
    LOADING,
    LOADING_OK,
    LOADING_ERROR,
}

const route = useRoute();
const playlistId = route.params.playlistId as string;

const state = ref(State.LOADING);
const playlist = ref<Playlist | null>(null);

(async () => {
    try {
        state.value = State.LOADING;
        playlist.value = await api.getPlaylist(playlistId);
        state.value = State.LOADING_OK;
    } catch (e) {
        state.value = State.LOADING_ERROR;
        console.error(e);
    }
})();
</script>

<template>
    <template v-if="state == State.LOADING">
        Loading...
    </template>
    <template v-else-if="state == State.LOADING_ERROR">
        Failed to load playlist {{ playlistId }}
    </template>
    <template v-else-if="playlist">
        <h1>{{ playlist.Name }}</h1>

        <div v-for="track in playlist.Tracks">
            <span v-if="track.TapeTrack.Artist">{{ track.TapeTrack.Artist }} - </span>{{ track.TapeTrack.Title }}
        </div>
    </template>
    <template v-else>
        Unknown error
    </template>
</template>
