<script setup lang="ts">
import api, { type Album, type RelatedItems } from "@/api";
import { useRoute } from "vue-router";
import { computed, ref, toRaw } from "vue";
import router from "@/router";
import TapeGrid from "@/components/TapeGrid.vue";

enum State {
    LOADING,
    LOADING_OK,
    LOADING_ERROR,
    SAVING,
    SAVING_OK,
    SAVING_ERROR,
    DELETING,
    DELETING_OK,
    DELETING_ERROR,
}

const route = useRoute();
const albumId = route.params.albumId as string;

const state = ref(State.LOADING);

const album = ref<Album | null>(null);
const editedAlbum = ref<Album | null>(null);

const releaseDate = computed({
    get(): string | null {
        const val = editedAlbum.value?.ReleaseDate;
        if (val == null) {
            return null;
        }

        const match = val.match(/^(\d{4}-\d{2}-\d{2})T/);
        if (match == null) {
            return null;
        }

        return `${match[1]}`;
    },
    set(val: string | null) {
        const album = editedAlbum.value;
        if (album == null) {
            return;
        }

        if (val == null) {
            album.ReleaseDate = null;
        } else {
            const match = val.match(/^(\d{4}-\d{2}-\d{2})$/);
            if (match != null) {
                album.ReleaseDate = `${match[1]}T00:00:00Z`;
            }
        }
    },
});

const relatedItems = ref<RelatedItems | null>(null);

const isEdited = computed(() => {
    return JSON.stringify(album.value) != JSON.stringify(editedAlbum.value);
});

const isBusy = computed(() => {
    switch (state.value) {
        case State.LOADING:
        case State.SAVING:
        case State.DELETING:
            return true;
        default:
            return false;
    }
});

function reset() {
    editedAlbum.value = structuredClone(toRaw(album.value));
}

async function removeReleaseDate() {
    releaseDate.value = null;
}

async function saveAlbum() {
    try {
        state.value = State.SAVING;
        album.value = await api.updateAlbum(albumId, editedAlbum.value!);
        state.value = State.SAVING_OK;

        reset();
    } catch (e) {
        state.value = State.SAVING_ERROR;
        console.error(e);
    }
}

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

    reset();
})();
</script>

<template>
    <template v-if="state == State.LOADING">
        Loading...
    </template>
    <template v-else-if="state == State.LOADING_ERROR">
        Failed to load album {{ albumId }}
    </template>
    <template v-else-if="editedAlbum">
        <h1>{{ editedAlbum.Name }}</h1>
        <h2>by {{ editedAlbum.Artist }}</h2>

        <div>
            <button :disabled="!isEdited || isBusy" @click="reset">Reset</button>
            <button :disabled="!isEdited || isBusy" @click="saveAlbum">Save</button>
            <button :disabled="isBusy" @click="deleteAlbum">Delete</button>
        </div>
        <div v-if="state == State.SAVING">Saving...</div>
        <div v-else-if="state == State.SAVING_OK">Saved</div>
        <div v-else-if="state == State.SAVING_ERROR">Failed to save the album</div>
        <div v-else-if="state == State.DELETING">Deleting...</div>
        <div v-else-if="state == State.DELETING_OK">Deleted</div>
        <div v-else-if="state == State.DELETING_ERROR">Failed to delete the album</div>

        <hr>

        <input type="date" v-model="releaseDate">
        <button :disabled="isBusy || releaseDate == null" @click="removeReleaseDate">Remove release date</button>

        <hr>

        <div v-for="track in editedAlbum.Tracks">
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
