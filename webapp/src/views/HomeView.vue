<script setup lang="ts">
import { ref } from "vue";

enum State {
  WAITING,
  SUBMITTING,
  ERROR,
  SUCCESS,
}

interface HistoryEntry {
  url: string
  result: State
  error: string | null
}

const urlInput = ref("");
const state = ref(State.WAITING);
const history = ref<HistoryEntry[]>([]);

async function submit() {
  const url = urlInput.value;

  state.value = State.SUBMITTING;

  const response = await fetch(
    "/api/import?" + new URLSearchParams({ url, format: "ba" }),
    { method: "POST" },
  );

  if (response.ok) {
    state.value = State.SUCCESS;
    urlInput.value = "";

    history.value.splice(0, 0, {
      url,
      result: State.SUCCESS,
      error: null,
    });
  } else {
    state.value = State.ERROR;

    history.value.splice(0, 0, {
      url,
      result: State.ERROR,
      error: `${response.status} ${response.statusText}`,
    });
  }
}

</script>

<template>
  <div>
    <input type="url" v-model="urlInput" :disabled="state == State.SUBMITTING">
    <button @click="submit()" :disabled="state == State.SUBMITTING">Import</button>
  </div>
  <div v-if="state == State.SUBMITTING">Importing...</div>
  <div>
    <div v-for="it in history">
      <a :href="it.url">{{ it.url }}</a>: {{ State[it.result] }} {{ it.error }}
    </div>
  </div>
</template>
