<script setup lang="ts">
import api, { type Playlist } from "@/api";
import { useRoute } from "vue-router";
import { computed, ref } from "vue";
import router from "@/router";

enum State {
    LOADING,
    LOADING_OK,
    LOADING_ERROR,
    DELETING,
    DELETING_OK,
    DELETING_ERROR,
}

const route = useRoute();
const playlistId = route.params.playlistId as string;

const state = ref(State.LOADING);
const playlist = ref<Playlist | null>(null);

const isBusy = computed(() => {
    switch (state.value) {
        case State.LOADING:
        case State.DELETING:
            return true;
        default:
            return false;
    }
});

async function deletePlaylist() {
    try {
        state.value = State.DELETING;
        await api.deletePlaylist(playlistId);
        state.value = State.DELETING_OK;

        router.push({ name: "home" });
    } catch (e) {
        state.value = State.DELETING_ERROR;
        console.error(e);
    }
}

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

        <h2>
            <div>
                <button :disabled="isBusy" @click="deletePlaylist">Delete</button>
            </div>
            <div v-if="state == State.DELETING_ERROR">Failed to delete the playlist</div>
        </h2>

        <hr>

        <div v-for="track in playlist.Tracks">
            <span v-if="track.TapeTrack.Artist">{{ track.TapeTrack.Artist }} - </span>{{ track.TapeTrack.Title }}
        </div>
    </template>
    <template v-else>
        Unknown error
    </template>
</template>
