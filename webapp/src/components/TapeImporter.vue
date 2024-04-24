<script setup lang="ts">
import api, { type ImportQueueItem } from "@/api";
import { computed, ref } from "vue";
import TapeImporterQueueItem from "@/components/TapeImporterQueueItem.vue";

enum State {
    LOADING,
    LOADING_OK,
    LOADING_ERROR,
    SUBMITTING,
    SUBMITTING_OK,
    SUBMITTING_ERROR,
}

const urlInput = ref("");
const state = ref(State.LOADING);

const queue = ref<ImportQueueItem[]>([]);
const isFormDisabled = computed(() => state.value == State.LOADING || state.value == State.SUBMITTING);

async function refreshQueue() {
    try {
        state.value = State.LOADING;

        queue.value = await api.getImportQueue();

        state.value = State.LOADING_OK;
    } catch (e) {
        state.value = State.LOADING_ERROR;
        console.error(e);
    }
}

async function submit() {
    const url = urlInput.value;

    try {
        state.value = State.SUBMITTING;

        const response = await api.addToImportQueue(url);
        queue.value.splice(0, 0, response);
        urlInput.value = "";

        state.value = State.SUBMITTING_OK;
    } catch (e) {
        state.value = State.SUBMITTING_ERROR;
        console.error(e);
    }
}

function onDeleted(id: string) {
    const index = queue.value.findIndex(it => it.Id == id);
    if (index >= 0) {
        queue.value.splice(index, 1);
    }
}

refreshQueue();
</script>

<template>
    <div :disabled="state == State.LOADING || state == State.SUBMITTING">
        <input type="url" v-model="urlInput" :disabled="isFormDisabled">
        <button @click="submit()" :disabled="isFormDisabled">Import</button>
    </div>

    <div v-if="state == State.LOADING">Loading the import queue...</div>
    <div v-else-if="state == State.LOADING_ERROR">Failed to load the import queue</div>
    <div v-else-if="state == State.SUBMITTING">Adding the URL to the import queue...</div>
    <div v-else-if="state == State.SUBMITTING_ERROR">Failed to add the URL to the import queue</div>

    <div>
        <TapeImporterQueueItem v-for="item in queue" :item="item" @deleted="onDeleted(item.Id)" />
    </div>
</template>
