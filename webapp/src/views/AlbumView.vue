<script setup lang="ts">
import api, { type Album, type RelatedItems } from "@/api";
import { useRoute } from "vue-router";
import { computed, ref } from "vue";
import router from "@/router";
import TapeGrid from "@/components/TapeGrid.vue";

enum State {
    LOADING,
    LOADING_OK,
    LOADING_ERROR,
    DELETING,
    DELETING_OK,
    DELETING_ERROR,
}

const route = useRoute();
const albumId = route.params.albumId as string;

const state = ref(State.LOADING);
const album = ref<Album | null>(null);

const relatedItems = ref<RelatedItems | null>(null);

const isBusy = computed(() => {
    switch (state.value) {
        case State.LOADING:
        case State.DELETING:
            return true;
        default:
            return false;
    }
});

async function deleteAlbum() {
    try {
        state.value = State.DELETING;
        await api.deleteAlbum(albumId);
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

        const albumAsync = api.getAlbum(albumId);
        const relatedItemsAsync = api.getAlbumRelationships(albumId);

        album.value = await albumAsync;
        relatedItems.value = await relatedItemsAsync;

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
        Failed to load album {{ albumId }}
    </template>
    <template v-else-if="album">
        <h1>{{ album.Name }}</h1>
        <h2>by {{ album.Artist }}</h2>

        <h2>
            <div>
                <button :disabled="isBusy" @click="deleteAlbum">Delete</button>
            </div>
            <div v-if="state == State.DELETING_ERROR">Failed to delete the album</div>
        </h2>

        <hr>

        <div v-for="track in album.Tracks">
            <span v-if="track.TapeTrack.Artist">{{ track.TapeTrack.Artist }} - </span>{{ track.TapeTrack.Title }}
        </div>

        <template v-if="relatedItems?.Tapes">
            <hr>

            <h2>Linked tapes</h2>
            <TapeGrid v-model="relatedItems.Tapes" />
        </template>
    </template>
    <template v-else>
        Unknown error
    </template>
</template>
