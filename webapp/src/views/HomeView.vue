<script setup lang="ts">
import { ref } from "vue";
import tapeApi, { type TapeListRs } from "@/api/tapes";
import TapeGrid from "@/components/TapeGrid.vue";

enum State {
    LOADING,
    OK,
    ERROR,
}

const state = ref(State.LOADING);
const tapes = ref<TapeListRs[]>([]);

(async () => {
    try {
        state.value = State.LOADING;

        const tapesAsync = tapeApi.listTapes();

        tapes.value = await tapesAsync;

        state.value = State.OK;
    } catch (e) {
        state.value = State.ERROR;
    }
})();
</script>

<template>
    <div v-if="state == State.LOADING">Loading...</div>
    <div v-else-if="state == State.ERROR">Failed to load tapes</div>
    <div v-else-if="state == State.OK">
        <TapeGrid :model-value="tapes" />
    </div>
    <div v-else>Unknown state {{ state }}</div>
</template>
