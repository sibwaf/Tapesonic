<script setup lang="ts">
import api, { type Tape } from "@/api";
import { useRoute } from "vue-router";
import { computed, ref, toRaw } from "vue";
import TapeTrackListEditor from "@/components/TapeTrackListEditor.vue";

enum State {
    LOADING,
    LOADING_OK,
    LOADING_ERROR,
    SAVING,
    SAVING_OK,
    SAVING_ERROR,
}

const route = useRoute();
const tapeId = route.params.tapeId as string;

const state = ref(State.LOADING);
const tape = ref<Tape | null>(null);
const editedTape = ref<Tape | null>(null);

function msToTimestamp(ms: number): string {
    const seconds = (ms / 1000) % 60;
    const minutes = (ms / 1000 / 60) % 60;
    const hours = (ms / 1000 / 60 / 60);

    function format(n: number): string {
        const value = Math.floor(n);
        return value < 10 ? `0${value}` : value.toString();
    }

    return `${format(hours)}:${format(minutes)}:${format(seconds)}`;
}

const isEdited = computed(() => {
    return JSON.stringify(tape.value) != JSON.stringify(editedTape.value);
});

function reset() {
    editedTape.value = structuredClone(toRaw(tape.value));
}

async function save() {
    try {
        state.value = State.SAVING;
        await api.saveTape(tapeId, editedTape.value!);
        tape.value = structuredClone(toRaw(editedTape.value));
        state.value = State.SAVING_OK;
    } catch (e) {
        state.value = State.SAVING_ERROR;
        console.error(e);
    }
}

(async () => {
    try {
        state.value = State.LOADING;
        tape.value = await api.getTape(tapeId);
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
        Failed to load tape {{ tapeId }}
    </template>
    <template v-else-if="editedTape">
        <h1>{{ editedTape.Name }}</h1>
        <h2>by {{ editedTape.AuthorName }}</h2>
        <TapeTrackListEditor v-if="editedTape" v-model="editedTape.Tracks" />

        <button :disabled="!isEdited && state != State.SAVING" @click="reset">Reset</button>
        <button :disabled="!isEdited && state != State.SAVING" @click="save">Save</button>

        <div v-if="state == State.SAVING">Saving...</div>
        <div v-else-if="state == State.SAVING_OK">Saved</div>
        <div v-else-if="state == State.SAVING_ERROR">Failed to save</div>
    </template>
    <template v-else>
        Unknown error
    </template>
</template>
