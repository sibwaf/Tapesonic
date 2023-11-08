<script setup lang="ts">
import api from "@/api";
import { ref } from "vue";

enum State {
    WAITING,
    SUBMITTING,
    ERROR,
    SUCCESS,
}

const urlInput = ref("");
const state = ref(State.WAITING);
const stateDescription = ref("");

async function submit() {
    const url = urlInput.value;

    state.value = State.SUBMITTING;
    stateDescription.value = "";

    const response = await api.import(url, "ba");

    if (response.Ok) {
        state.value = State.SUCCESS;
        urlInput.value = "";
        stateDescription.value = url;
    } else {
        state.value = State.ERROR;
        stateDescription.value = response.Error ?? "Unknown error";
    }
}
</script>

<template>
    <div>
        <input type="url" v-model="urlInput" :disabled="state == State.SUBMITTING">
        <button @click="submit()" :disabled="state == State.SUBMITTING">Import</button>
    </div>
    <div v-if="state == State.SUBMITTING">Importing...</div>
    <div v-else-if="state == State.SUCCESS">Imported {{ stateDescription }}</div>
    <div v-else-if="state == State.ERROR">Failed to import: {{ stateDescription }}</div>
</template>
