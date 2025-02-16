<script lang="ts" setup>
import api, { type TrackRs } from '@/api';
import { computed, ref, toRaw } from 'vue';

enum State {
    WAITING,
    SEARCHING,
    SEARCH_DONE,
    SEARCH_FAILED,
    IMPORTING,
    IMPORT_DONE,
    IMPORT_FAILED,
}

const emit = defineEmits<{
    (e: "add-track", track: TrackRs): void
}>();

const state = ref(State.WAITING);
const isBusy = computed(() => state.value == State.SEARCHING || state.value == State.IMPORTING);

const query = ref("");

const searchResult = ref<TrackRs[]>([]);

async function search() {
    try {
        state.value = State.SEARCHING;

        searchResult.value = [];
        searchResult.value = await api.searchTracks(query.value);

        state.value = State.SEARCH_DONE;
    } catch (e) {
        state.value = State.SEARCH_FAILED;
        console.error(e);
    }
}

async function importSource() {
    try {
        state.value = State.IMPORTING;

        const result = await api.addSource(query.value);
        state.value = State.IMPORT_DONE;

        query.value = result.Url;
        await search();
    } catch (e) {
        state.value = State.IMPORT_FAILED;
        console.error(e);
    }
}

function onAdd(track: TrackRs) {
    emit("add-track", toRaw(track));
}

function onAddAll() {
    for (const track of searchResult.value) {
        onAdd(track);
    }
}
</script>

<template>
    <div>
        <div>
            <input type="text" v-model="query" :disabled="isBusy">
            <button @click="search" :disabled="isBusy">Search</button>
            <button @click="importSource" :disabled="isBusy">Import</button>
        </div>
        <template v-if="state == State.SEARCHING">Searching...</template>
        <template v-else-if="state == State.SEARCH_DONE">
            <div v-if="searchResult.length > 0">
                <div v-for="track in searchResult" :key="track.Id">
                    <button @click="onAdd(track)">Add</button> {{ track.Artist }} - {{ track.Title }}
                </div>
                <button @click="onAddAll">Add all</button>
            </div>
            <div v-else>
                Nothing found, try importing?
            </div>
        </template>
        <template v-else-if="state == State.SEARCH_FAILED">
            <div>Got error while searching</div>
        </template>
        <template v-else-if="state == State.IMPORTING">Importing...</template>
        <template v-else-if="state == State.IMPORT_FAILED">
            <div>Got error while importing</div>
        </template>
    </div>
</template>
