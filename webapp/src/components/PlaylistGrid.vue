<script setup lang="ts">
import api, { type Playlist } from "@/api";
import { ref } from "vue";
import { RouterLink } from "vue-router";

enum State {
    LOADING, OK, ERROR,
}

const state = ref(State.LOADING);
const playlists = ref<Playlist[]>([]);

(async () => {
    try {
        playlists.value = await api.getAllPlaylists();
        state.value = State.OK;
    } catch (e) {
        state.value = State.ERROR;
    }
})();
</script>

<template>
    <div v-if="state == State.OK">
        <RouterLink v-for="playlist in playlists" :key="playlist.Id" :to="'/playlists/' + playlist.Id">
            <div>{{ playlist.Name }}</div>
        </RouterLink>
    </div>
    <div v-else-if="state == State.LOADING">
        Loading...
    </div>
    <div v-else>
        Failed to load the playlist grid
    </div>
</template>
