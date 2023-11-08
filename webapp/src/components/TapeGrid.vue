<script setup lang="ts">
import api, { type Tape } from "@/api";
import { ref } from "vue";
import { RouterLink } from "vue-router";

enum State {
    LOADING, OK, ERROR,
}

const state = ref(State.LOADING);
const tapes = ref<Tape[]>([]);

(async () => {
    try {
        tapes.value = await api.getAllTapes();
        state.value = State.OK;
    } catch (e) {
        state.value = State.ERROR;
    }
})();
</script>

<template>
    <div v-if="state == State.OK">
        <RouterLink v-for="tape in tapes" :key="tape.Id" :to="'/tapes/' + tape.Id">
            <div>{{ tape.Name }}</div>
            <div>by {{ tape.AuthorName }}</div>
        </RouterLink>
    </div>
    <div v-else-if="state == State.LOADING">
        Loading...
    </div>
    <div v-else>
        Failed to load the tape grid
    </div>
</template>
