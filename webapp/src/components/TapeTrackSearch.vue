<script lang="ts" setup>
import tracksApi, { type TrackRs } from '@/api/tracks';
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

        const rs = await tracksApi.searchTracks(query.value);
        searchResult.value = rs.filter(it => it.SourceId != ""); // todo: only non-remote tracks can be used in tapes for now

        state.value = State.SEARCH_DONE;
    } catch (e) {
        state.value = State.SEARCH_FAILED;
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
            <button @click="search" :disabled="isBusy">Search or import</button>
        </div>
        <template v-if="state == State.SEARCHING">Searching...</template>
        <template v-else-if="state == State.SEARCH_DONE">
            <div v-if="searchResult.length > 0">
                <div v-for="track in searchResult" :key="track.Id">
                    <button @click="onAdd(track)">Add</button>
                    <span>
                        <span v-if="track.Artist">{{ track.Artist.Name }}</span>
                        <span v-if="track.Artist && track.Title">&ensp;-&ensp;</span>
                        <span v-if="track.Title">{{ track.Title }}</span>
                    </span>
                </div>
                <button @click="onAddAll">Add all</button>
            </div>
            <div v-else>
                Nothing found
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
